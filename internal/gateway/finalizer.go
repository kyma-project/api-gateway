package gateway

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const kymaGatewayFinalizer string = "gateways.operator.kyma-project.io/kyma-gateway"

func hasKymaGatewayFinalizer(apiGatewayCR v1alpha1.APIGateway) bool {
	return controllerutil.ContainsFinalizer(&apiGatewayCR, kymaGatewayFinalizer)
}

func addKymaGatewayFinalizer(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) error {
	ctrl.Log.Info("Adding finalizer", "finalizer", kymaGatewayFinalizer)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); err != nil {
			return err
		}

		controllerutil.AddFinalizer(apiGatewayCR, kymaGatewayFinalizer)

		if err := k8sClient.Update(ctx, apiGatewayCR); err != nil {
			return err
		}

		return nil
	})
}

func removeKymaGatewayFinalizer(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) error {
	ctrl.Log.Info("Removing finalizer", "finalizer", kymaGatewayFinalizer)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); err != nil {
			return err
		}

		controllerutil.RemoveFinalizer(apiGatewayCR, kymaGatewayFinalizer)

		if err := k8sClient.Update(ctx, apiGatewayCR); err != nil {
			return err
		}

		return nil
	})
}
