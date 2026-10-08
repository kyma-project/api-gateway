package gateway

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/api-gateway/internal/controller"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/resources"
)

const (
	nonGardenerDomainName = "local.kyma.dev"
)

var checkDefaultGatewayReference = func(ctx context.Context, c client.Client, res resources.Resource) bool {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(res.GVK)

	err := c.Get(ctx, client.ObjectKey{
		Namespace: res.Namespace,
		Name:      res.Name,
	}, u)

	if err != nil {
		ctrl.Log.Error(err, "Error happened during getting object")
	}

	if res.GVK.Kind == "APIRule" && u.Object["spec"] != nil {
		return u.Object["spec"].(map[string]any)["gateway"] == KymaGatewayFullName
	} else if res.GVK.Kind == "VirtualService" && u.Object["spec"] != nil {
		gateways := u.Object["spec"].(map[string]any)["gateways"]
		if gateways != nil {
			for _, gateway := range gateways.([]any) {
				if gateway == KymaGatewayFullName {
					return true
				}
			}
		}
	}

	return false
}

// ReconcileKymaGateway reconciles the kyma-gateway and creates all required resources for the Gateway to fully work. It also adds a finalizer to
// APIGateway CR and handles the deletion of the resources if the APIGateway CR is deleted.
func ReconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway, apiGatewayResourceListPath string) controller.Status {
	ctrl.Log.Info("Reconcile Kyma Gateway", "enabled", apiGatewayCR.Spec.EnableKymaGateway)
	if isKymaGatewayEnabled(*apiGatewayCR) && !apiGatewayCR.IsInDeletion() && !hasKymaGatewayFinalizer(*apiGatewayCR) {
		if err := addKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			msg := "Failed to add finalizer during Kyma Gateway reconciliation"
			return controller.ErrorStatus(err, msg,
				v1alpha1.ErrorCondition(v1alpha1.ReasonReconcileFailed, msg),
				v1alpha1.KymaGatewayErrorCondition(msg))
		}
	}

	if !hasKymaGatewayFinalizer(*apiGatewayCR) {
		ctrl.Log.Info("There is no Kyma Gateway finalizer, skipping reconciliation")
		return controller.ReadyStatus(v1alpha1.ReadyCondition(), v1alpha1.KymaGatewayReadyCondition())
	}

	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		resourceFinder, err := resources.NewResourcesFinderFromConfigYaml(ctx, k8sClient, ctrl.Log, apiGatewayResourceListPath)
		if err != nil {
			msg := "Could not read customer resources finder configuration"
			return controller.ErrorStatus(err, msg,
				v1alpha1.ErrorCondition(v1alpha1.ReasonReconcileFailed, msg),
				v1alpha1.KymaGatewayErrorCondition(msg))
		}

		clientResources, err := resourceFinder.FindUserCreatedResources(checkDefaultGatewayReference)
		if err != nil {
			msg := "Could not get customer resources from the cluster"
			return controller.ErrorStatus(err, msg,
				v1alpha1.ErrorCondition(v1alpha1.ReasonReconcileFailed, msg),
				v1alpha1.KymaGatewayErrorCondition(msg))
		}

		if len(clientResources) > 0 {
			var blockingResources []string
			for _, res := range clientResources {
				ctrl.Log.Info("Custom resource is blocking Kyma Gateway deletion", "gvk", res.GVK.String(), "namespace", res.Namespace, "name", res.Name)
				if len(blockingResources) < 5 {
					blockingResources = append(blockingResources, res.Name)
				}
			}

			msg := "Kyma Gateway deletion blocked because of the existing custom resources: " + strings.Join(blockingResources, ", ")
			return controller.WarningStatus(
				fmt.Errorf("could not delete Kyma Gateway since there are %d custom resource(s) present that block its deletion", len(clientResources)),
				"There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				v1alpha1.DeletionBlockedExistingResourcesCondition(msg),
				v1alpha1.KymaGatewayDeletionBlockedCondition(msg))
		}
	}

	conditions, err := reconcile(ctx, k8sClient, *apiGatewayCR)
	if errors.Is(err, ErrCertificatePending) && apiGatewayCR.Status.State != v1alpha1.Ready {
		return controller.ProcessingStatus(append([]metav1.Condition{v1alpha1.KymaGatewayProcessingCondition()}, conditions...)...)
	}
	if err != nil {
		msg := "Error during Kyma Gateway reconciliation: " + err.Error()
		return controller.ErrorStatus(err, msg, append([]metav1.Condition{v1alpha1.ErrorCondition(v1alpha1.ReasonReconcileFailed, msg), v1alpha1.KymaGatewayErrorCondition(msg)}, conditions...)...)
	}

	// Besides on disabling the Kyma gateway, we also need to remove the finalizer on APIGateway deletion to make sure we are not blocking the deletion of the CR.
	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		if err := removeKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			msg := "Failed to remove finalizer during Kyma Gateway reconciliation"
			return controller.ErrorStatus(err, msg, append([]metav1.Condition{v1alpha1.ErrorCondition(v1alpha1.ReasonReconcileFailed, msg), v1alpha1.KymaGatewayErrorCondition(msg)}, conditions...)...)
		}
	}

	return controller.ReadyStatus(append([]metav1.Condition{v1alpha1.ReadyCondition(), v1alpha1.KymaGatewayReadyCondition()}, conditions...)...)
}

func reconcile(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) ([]metav1.Condition, error) {
	domain, err := reconciliations.GetGardenerDomain(ctx, k8sClient)
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}
	if domain == "" {
		domain = nonGardenerDomainName
	}

	_, err = dependencies.Gardener().AreAvailable(ctx, k8sClient)
	onGardener := err == nil && domain != nonGardenerDomainName

	conditions := make([]metav1.Condition, 0, 2)

	if onGardener {
		if err := reconcileKymaGatewayDnsEntry(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return append(conditions, v1alpha1.DNSEntryErrorCondition(err.Error())), err
		}
		conditions = append(conditions, v1alpha1.DNSEntryReadyCondition())

		if err := reconcileKymaGatewayCertificate(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return append(conditions, v1alpha1.CertificateErrorCondition(err.Error())), err
		}
	} else {
		if err := reconcileNonGardenerCertificateSecret(ctx, k8sClient, apiGatewayCR); err != nil {
			return append(conditions, v1alpha1.CertificateErrorCondition(err.Error())), err
		}
	}

	if err := reconcileKymaGatewayVirtualService(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return conditions, err
	}

	if err := reconcilev1beta1andv2alpha1UIDeletion(ctx, k8sClient); err != nil {
		return conditions, err
	}

	if err := reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return conditions, err
	}

	// On Gardener the gateway depends on a Certificate that is issued asynchronously,
	// so we surface the certificate-specific condition here.
	if onGardener && isKymaGatewayEnabled(apiGatewayCR) && !apiGatewayCR.IsInDeletion() {
		if err := verifyKymaGatewayCertificateReady(ctx, k8sClient); err != nil {
			if errors.Is(err, ErrCertificatePending) {
				conditions = append(conditions, v1alpha1.CertificateProcessingCondition())
			} else {
				conditions = append(conditions, v1alpha1.CertificateErrorCondition(err.Error()))
			}
			return conditions, err
		}

		conditions = append(conditions, v1alpha1.CertificateReadyCondition())
	}

	return conditions, nil
}
