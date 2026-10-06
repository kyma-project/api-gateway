package gateway

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

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

type GatewayResult struct {
	State       v1alpha1.State
	Description string
	Conditions  []metav1.Condition
	Err         error
}

// ReconcileKymaGateway reconciles the kyma-gateway and creates all required resources for the Gateway to fully work. It also adds a finalizer to
// APIGateway CR and handles the deletion of the resources if the APIGateway CR is deleted.
func ReconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway, apiGatewayResourceListPath string) GatewayResult {
	ctrl.Log.Info("Reconcile Kyma Gateway", "enabled", apiGatewayCR.Spec.EnableKymaGateway)
	if isKymaGatewayEnabled(*apiGatewayCR) && !apiGatewayCR.IsInDeletion() && !hasKymaGatewayFinalizer(*apiGatewayCR) {
		if err := addKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			return GatewayResult{
				State:       v1alpha1.Error,
				Description: "Failed to add finalizer during Kyma Gateway reconciliation",
				Conditions:  []metav1.Condition{v1alpha1.KymaGatewayErrorCondition("Failed to add finalizer during Kyma Gateway reconciliation")},
				Err:         err,
			}
		}
	}

	if !hasKymaGatewayFinalizer(*apiGatewayCR) {
		ctrl.Log.Info("There is no Kyma Gateway finalizer, skipping reconciliation")
		return GatewayResult{
			State:       v1alpha1.Ready,
			Description: "",
			Conditions:  []metav1.Condition{v1alpha1.KymaGatewayReadyCondition()},
			Err:         nil,
		}
	}

	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		resourceFinder, err := resources.NewResourcesFinderFromConfigYaml(ctx, k8sClient, ctrl.Log, apiGatewayResourceListPath)
		if err != nil {
			return GatewayResult{
				State:       v1alpha1.Error,
				Description: "Could not read customer resources finder configuration",
				Conditions:  []metav1.Condition{v1alpha1.KymaGatewayErrorCondition("Could not read customer resources finder configuration")},
				Err:         err,
			}
		}

		clientResources, err := resourceFinder.FindUserCreatedResources(checkDefaultGatewayReference)
		if err != nil {
			return GatewayResult{
				State:       v1alpha1.Error,
				Description: "Could not get customer resources from the cluster",
				Conditions:  []metav1.Condition{v1alpha1.KymaGatewayErrorCondition("Could not get customer resources from the cluster")},
				Err:         err,
			}
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
			return GatewayResult{
				State:       v1alpha1.Warning,
				Description: "There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				Conditions:  []metav1.Condition{v1alpha1.KymaGatewayDeletionBlockedCondition(msg)},
				Err:         fmt.Errorf("could not delete Kyma Gateway since there are %d custom resource(s) present that block its deletion", len(clientResources)),
			}
		}
	}

	conditions, err := reconcile(ctx, k8sClient, *apiGatewayCR)
	if errors.Is(err, ErrCertificatePending) && apiGatewayCR.Status.State != v1alpha1.Ready {
		return GatewayResult{
			State:       v1alpha1.Processing,
			Description: "Kyma Gateway certificate pending",
			Conditions:  append([]metav1.Condition{v1alpha1.KymaGatewayProcessingCondition()}, conditions...),
			Err:         nil,
		}
	}
	if err != nil {
		msg := "Error during Kyma Gateway reconciliation: " + err.Error()
		return GatewayResult{
			State:       v1alpha1.Error,
			Description: msg,
			Conditions:  append([]metav1.Condition{v1alpha1.KymaGatewayErrorCondition(msg)}, conditions...),
			Err:         err,
		}
	}

	// Besides on disabling the Kyma gateway, we also need to remove the finalizer on APIGateway deletion to make sure we are not blocking the deletion of the CR.
	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		if err := removeKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			return GatewayResult{
				State:       v1alpha1.Error,
				Description: "Failed to remove finalizer during Kyma Gateway reconciliation",
				Conditions:  append([]metav1.Condition{v1alpha1.KymaGatewayErrorCondition("Failed to remove finalizer during Kyma Gateway reconciliation")}, conditions...),
				Err:         err,
			}
		}
	}

	return GatewayResult{
		State:       v1alpha1.Ready,
		Description: "",
		Conditions:  append([]metav1.Condition{v1alpha1.KymaGatewayReadyCondition()}, conditions...),
		Err:         nil,
	}
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
			return conditions, err
		}
		conditions = append(conditions, v1alpha1.DNSEntryReadyCondition())

		if err := reconcileKymaGatewayCertificate(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return conditions, err
		}
	} else {
		if err := reconcileNonGardenerCertificateSecret(ctx, k8sClient, apiGatewayCR); err != nil {
			return conditions, err
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
			}
			return conditions, err
		}

		conditions = append(conditions, v1alpha1.CertificateReadyCondition())
	}

	return conditions, nil
}
