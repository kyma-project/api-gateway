package api_gateway

import (
	"context"
	"fmt"

	"github.com/kyma-project/api-gateway/controllers"

	"k8s.io/client-go/util/retry"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ApiGatewayReconciliation interface {
	Reconcile(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) controllers.Status
}

type Reconciliation struct {
	Client client.Client
}

const (
	ApiGatewayFinalizer string = "gateways.operator.kyma-project.io/api-gateway"
)

func (i *Reconciliation) Reconcile(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) controllers.Status {
	ctrl.Log.Info("Reconcile API-Gateway CR")
	if !apiGatewayCR.IsInDeletion() && !hasFinalizer(apiGatewayCR) {
		controllerutil.AddFinalizer(apiGatewayCR, ApiGatewayFinalizer)
		if err := i.Client.Update(ctx, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Failed to add API-Gateway CR finalizer")
			return controllers.ErrorStatus(err, "Could not add API-Gateway CR finalizer")
		}
	}

	if !hasFinalizer(apiGatewayCR) {
		ctrl.Log.Info("API-Gateway CR has no finalizer, reconciliation is skipped.")
		return controllers.ReadyStatus()
	}

	if apiGatewayCR.IsInDeletion() {
		apiRulesCount, err := apiRulesCount(ctx, i.Client)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during obtaining existing APIRules")
		}

		if apiRulesCount > 0 {
			return controllers.WarningStatus(fmt.Errorf("could not delete API-Gateway module instance since there are %d APIRule(s) present that block its deletion", apiRulesCount),
				"There are APIRule(s) that block the deletion. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning")
		}

		if err := removeFinalizer(ctx, i.Client, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Error happened during API-Gateway CR finalizer removal")
			return controllers.ErrorStatus(err, "Could not remove finalizer")
		}
	}

	return controllers.ReadyStatus()
}

func apiRulesCount(ctx context.Context, k8sClient client.Client) (int, error) {
	apiRuleList := v1beta1.APIRuleList{}
	err := k8sClient.List(ctx, &apiRuleList)
	if err != nil {
		return 0, err
	}

	return len(apiRuleList.Items), nil
}

func hasFinalizer(apiGatewayCR *operatorv1alpha1.APIGateway) bool {
	return controllerutil.ContainsFinalizer(apiGatewayCR, ApiGatewayFinalizer)
}

func removeFinalizer(ctx context.Context, apiClient client.Client, apiGatewayCR *operatorv1alpha1.APIGateway) error {
	ctrl.Log.Info("Removing API-Gateway CR finalizer")
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := apiClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); getErr != nil {
			return getErr
		}

		controllerutil.RemoveFinalizer(apiGatewayCR, ApiGatewayFinalizer)
		if updateErr := apiClient.Update(ctx, apiGatewayCR); updateErr != nil {
			return updateErr
		}

		ctrl.Log.Info("Successfully removed API-Gateway CR finalizer")
		return nil
	})
}
