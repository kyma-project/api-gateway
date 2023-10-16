package gateway

import (
	"context"
	_ "embed"
	"fmt"

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
		ctrl.Log.Error(err, "Error happened during obtaining user created object")
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
			return controllers.ErrorStatus(err, "Failed to add finalizer during Kyma Gateway reconciliation")
		}
	}

	if !hasKymaGatewayFinalizer(*apiGatewayCR) {
		ctrl.Log.Info("There is no Kyma Gateway finalizer, skipping reconciliation")
		return controllers.ReadyStatus()
	}

	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		resourceFinder, err := resources.NewResourcesFinderFromConfigYaml(ctx, k8sClient, ctrl.Log, apiGatewayResourceListPath)
		if err != nil {
			return controllers.ErrorStatus(err, "Could not read customer resources finder configuration")
		}

		clientResources, err := resourceFinder.FindUserCreatedResources(checkDefaultGatewayReference)
		if err != nil {
			return controllers.ErrorStatus(err, "Could not get customer resources from the cluster")
		}

		if len(clientResources) > 0 {
			for _, res := range clientResources {
				ctrl.Log.Info("Custom resource is blocking Kyma Gateway deletion", res.GVK.String(), fmt.Sprintf("%s/%s", res.Namespace, res.Name))
			}

			return controllers.WarningStatus(fmt.Errorf("could not delete Kyma Gateway since there are %d custom resource(s) present that block its deletion", len(clientResources)),
				"There are custom resources that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning")
		}
	}

	isGardenerCluster, err := reconciliations.RunsOnGardenerCluster(ctx, k8sClient)
	if err != nil {
		return controllers.ErrorStatus(err, "Error during Kyma Gateway reconciliation")
	}

	var reconcileErr error
	if isGardenerCluster {
		reconcileErr = reconcileGardenerKymaGateway(ctx, k8sClient, *apiGatewayCR)
	} else {
		reconcileErr = reconcileNonGardenerKymaGateway(ctx, k8sClient, *apiGatewayCR)
	}

	if reconcileErr != nil {
		return controllers.ErrorStatus(reconcileErr, "Error during Kyma Gateway reconciliation")
	}

	// Besides on disabling the Kyma gateway, we also need to remove the finalizer on APIGateway deletion to make sure we are not blocking the deletion of the CR.
	if !isKymaGatewayEnabled(*apiGatewayCR) || apiGatewayCR.IsInDeletion() {
		if err := removeKymaGatewayFinalizer(ctx, k8sClient, apiGatewayCR); err != nil {
			return controllers.ErrorStatus(err, "Failed to remove finalizer during Kyma Gateway reconciliation")
		}
	}

	return controllers.ReadyStatus()
}

func reconcileGardenerKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	domain, err := reconciliations.GetGardenerDomain(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to get Kyma gateway domain: %v", err)
	}

	if err := reconcileKymaGatewayDnsEntry(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return err
	}

	if err := reconcileKymaGatewayCertificate(ctx, k8sClient, apiGatewayCR, domain); err != nil {
		return err
	}

	return reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, domain)
}

func reconcileNonGardenerKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	if err := reconcileNonGardenerCertificateSecret(ctx, k8sClient, apiGatewayCR); err != nil {
		return err
	}

	return reconcileKymaGateway(ctx, k8sClient, apiGatewayCR, nonGardenerDomainName)
}
