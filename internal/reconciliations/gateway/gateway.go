package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/internal/conditions"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		return u.Object["spec"].(map[string]interface{})["gateway"] == KymaGatewayFullName
	} else if res.GVK.Kind == "VirtualService" && u.Object["spec"] != nil {
		gateways := u.Object["spec"].(map[string]interface{})["gateways"]
		if gateways != nil {
			for _, gateway := range gateways.([]interface{}) {
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
// Returns a Status object with the result of the reconciliation and an error if the reconciliation failed.
func ReconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway, apiGatewayResourceListPath string) controllers.Status {
	ctrl.Log.Info("Reconcile Kyma Gateway", "enabled", apiGatewayCR.Spec.EnableKymaGateway)
	if isKymaGatewayEnabled(*apiGatewayCR) && !apiGatewayCR.IsInDeletion() && !hasKymaGatewayFinalizer(*apiGatewayCR) {
		if err := addKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			return controllers.ErrorStatus(err, "Failed to add finalizer during Kyma Gateway reconciliation", conditions.KymaGatewayReconcileFailed.Condition())
		}
	}

	if !hasKymaGatewayFinalizer(*apiGatewayCR) {
		ctrl.Log.Info("There is no Kyma Gateway finalizer, skipping reconciliation")
		return controllers.ReadyStatus(conditions.KymaGatewayReconcileSucceeded.Condition())
	}

	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		resourceFinder, err := resources.NewResourcesFinderFromConfigYaml(ctx, k8sClient, ctrl.Log, apiGatewayResourceListPath)
		if err != nil {
			return controllers.ErrorStatus(err, "Could not read customer resources finder configuration", conditions.KymaGatewayReconcileFailed.Condition())
		}

		clientResources, err := resourceFinder.FindUserCreatedResources(checkDefaultGatewayReference)
		if err != nil {
			return controllers.ErrorStatus(err, "Could not get customer resources from the cluster", conditions.KymaGatewayReconcileFailed.Condition())
		}

		if len(clientResources) > 0 {
			var blockingResources []string
			for _, res := range clientResources {
				ctrl.Log.Info("Custom resource is blocking Kyma Gateway deletion", "gvk", res.GVK.String(), "namespace", res.Namespace, "name", res.Name)
				if len(blockingResources) < 5 {
					blockingResources = append(blockingResources, res.Name)
				}
			}

			return controllers.WarningStatus(fmt.Errorf("could not delete Kyma Gateway since there are %d custom resource(s) present that block its deletion", len(clientResources)),
				"There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				conditions.KymaGatewayDeletionBlocked.AdditionalMessage(": "+strings.Join(blockingResources, ", ")).Condition())
		}
	}

	if err := reconcile(ctx, k8sClient, *apiGatewayCR); err != nil {
		return controllers.ErrorStatus(err, "Error during Kyma Gateway reconciliation", conditions.KymaGatewayReconcileFailed.Condition())
	}

	// Besides on disabling the Kyma gateway, we also need to remove the finalizer on APIGateway deletion to make sure we are not blocking the deletion of the CR.
	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		if err := removeKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			return controllers.ErrorStatus(err, "Failed to remove finalizer during Kyma Gateway reconciliation", conditions.KymaGatewayReconcileFailed.Condition())
		}
	}

	return controllers.ReadyStatus(conditions.KymaGatewayReconcileSucceeded.Condition())
}

func reconcile(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	domain, err := reconciliations.GetGardenerDomain(ctx, k8sClient)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	if domain == "" {
		domain = nonGardenerDomainName
	}
	if _, err := dependencies.Gardener().AreAvailable(ctx, k8sClient); err == nil && domain != nonGardenerDomainName {
		if err := reconcileKymaGatewayDnsEntry(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return err
		}

		if err := reconcileKymaGatewayCertificate(ctx, k8sClient, apiGatewayCR, domain); err != nil {
			return err
		}
	} else {
		if err := reconcileNonGardenerCertificateSecret(ctx, k8sClient, apiGatewayCR); err != nil {
			return err
		}
	}
	if err := reconcileKymaGatewayVirtualService(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return err
	}
	return reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, domain)
}
