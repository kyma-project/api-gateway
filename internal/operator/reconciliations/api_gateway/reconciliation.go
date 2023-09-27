package api_gateway

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/described_errors"
	"github.com/kyma-project/api-gateway/internal/operator/resources"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ApiGatewayReconciliation interface {
	Reconcile(ctx context.Context, apiGatewayCR operatorv1alpha1.APIGateway, apiGatewayResourceListPath string) (operatorv1alpha1.APIGateway, described_errors.DescribedError)
}

type Reconciliation struct {
	Client client.Client
}

const (
	reconciliationFinalizer string = "api-gateway.operator.kyma-project.io/api-gateway-reconciliation"
	dafaultGatewayName      string = "kyma-gateway"
	dafaultGatewayNS        string = "kyma-system"
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

	if res.GVK.Kind == "APIRule" {
		return u.Object["spec"].(map[string]interface{})["gateway"] == dafaultGatewayNS+"/"+dafaultGatewayName
	} else if res.GVK.Kind == "VirtualService" {
		gateways := u.Object["spec"].(map[string]interface{})["gateways"].([]interface{})
		for _, gateway := range gateways {
			if gateway == dafaultGatewayNS+"/"+dafaultGatewayName {
				return true
			}
		}
	}

	return false
}

func (i *Reconciliation) Reconcile(ctx context.Context, apiGatewayCR operatorv1alpha1.APIGateway, apiGatewayResourceListPath string) (operatorv1alpha1.APIGateway, described_errors.DescribedError) {
	if shouldDelete(apiGatewayCR) && hasReconciliationFinalizer(apiGatewayCR) {
		ctrl.Log.Info("Starting API-Gateway deletion")

		resourceFinder, err := resources.NewResourcesFinderFromConfigYaml(ctx, i.Client, ctrl.Log, apiGatewayResourceListPath)
		if err != nil {
			return apiGatewayCR, described_errors.NewDescribedError(err, "Could not read customer resources finder configuration")
		}

		clientResources, err := resourceFinder.FindUserCreatedResources(checkDefaultGatewayReference)
		if err != nil {
			return apiGatewayCR, described_errors.NewDescribedError(err, "Could not get customer resources from the cluster")
		}

		if len(clientResources) > 0 {
			for _, res := range clientResources {
				ctrl.Log.Info("Customer resource is blocking API-Gateway deletion", res.GVK.Kind, fmt.Sprintf("%s/%s", res.Namespace, res.Name))
			}

			return apiGatewayCR, described_errors.NewDescribedError(fmt.Errorf("could not delete API-Gateway module instance since there are %d custom resource(s) present that block its deletion", len(clientResources)),
				"There are custom resource(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning").DisableErrorWrap().SetWarning()
		}

		//Temporarily having this simple solution for deletion of Kyma default gateway before we migrate whole management of Kyma default gateway from Istio component
		if err := deleteDefaultGateway(ctx, i.Client); err != nil {
			ctrl.Log.Error(err, "Error happened during API-Gateway reconciliation on default gateway removal")
			return apiGatewayCR, described_errors.NewDescribedError(err, "Could not remove Kyma default gateway")
		}

		if err := removeReconciliationFinalizer(ctx, i.Client, &apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Error happened during API-Gateway reconciliation on finalizer removal")
			return apiGatewayCR, described_errors.NewDescribedError(err, "Could not remove finalizer")
		}

		ctrl.Log.Info("API-Gateway deletion completed")
	} else {
		if !hasReconciliationFinalizer(apiGatewayCR) {
			controllerutil.AddFinalizer(&apiGatewayCR, reconciliationFinalizer)
			if err := i.Client.Update(ctx, &apiGatewayCR); err != nil {
				ctrl.Log.Error(err, "Failed to add API-Gateway reconciliation finalizer")
				return apiGatewayCR, described_errors.NewDescribedError(err, "Could not add finalizer")
			}
		}
	}

	return apiGatewayCR, nil
}

// shouldDelete returns true when API-Gateway should be deleted
func shouldDelete(apiGatewayCR operatorv1alpha1.APIGateway) bool {
	return !apiGatewayCR.DeletionTimestamp.IsZero()
}

func hasReconciliationFinalizer(apiGatewayCR operatorv1alpha1.APIGateway) bool {
	return controllerutil.ContainsFinalizer(&apiGatewayCR, reconciliationFinalizer)
}

func removeReconciliationFinalizer(ctx context.Context, apiClient client.Client, apiGatewayCR *operatorv1alpha1.APIGateway) error {
	ctrl.Log.Info("Removing API-Gateway installation finalizer")
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := apiClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); getErr != nil {
			return getErr
		}

		controllerutil.RemoveFinalizer(apiGatewayCR, reconciliationFinalizer)
		if updateErr := apiClient.Update(ctx, apiGatewayCR); updateErr != nil {
			return updateErr
		}

		ctrl.Log.Info("Successfully removed API-Gateway installation finalizer")

		return nil
	})
}

func deleteDefaultGateway(ctx context.Context, client client.Client) error {
	ctrl.Log.Info("Delete Kyma default gateway", "name", dafaultGatewayName, "namespace", dafaultGatewayNS)
	return client.Delete(ctx, &networkingv1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dafaultGatewayName,
			Namespace: dafaultGatewayNS,
		},
	})
}
