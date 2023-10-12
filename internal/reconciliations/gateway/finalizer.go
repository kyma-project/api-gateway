package gateway

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
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
	controllerutil.AddFinalizer(apiGatewayCR, kymaGatewayFinalizer)
	return k8sClient.Update(ctx, apiGatewayCR)
}

func removeKymaGatewayFinalizer(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) error {
	ctrl.Log.Info("Removing finalizer", "finalizer", kymaGatewayFinalizer)
	controllerutil.RemoveFinalizer(apiGatewayCR, kymaGatewayFinalizer)
	return k8sClient.Update(ctx, apiGatewayCR)
}
