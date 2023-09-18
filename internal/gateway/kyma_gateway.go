package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kymaGatewayName      = "kyma-gateway"
	kymaGatewayNamespace = "kyma-system"
)

//go:embed kyma_gateway.yaml
var kymaGatewayManifest []byte

func reconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Kyma gateway", "KymaGatewayEnabled", isEnabled)

	if !isEnabled {
		return deleteKymaGateway(k8sClient)
	}

	templateValues := make(map[string]string)
	templateValues["Name"] = kymaGatewayName
	templateValues["Namespace"] = kymaGatewayNamespace
	templateValues["Domain"] = domain

	return reconcileResource(ctx, k8sClient, kymaGatewayManifest, templateValues)
}

func isKymaGatewayEnabled(cr v1alpha1.APIGateway) bool {
	return cr.Spec.EnableKymaGateway != nil && *cr.Spec.EnableKymaGateway == true
}

func deleteKymaGateway(k8sClient client.Client) error {
	ctrl.Log.Info("Deleting Kyma gateway if it exists")
	gw := v1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kymaGatewayName,
			Namespace: kymaGatewayNamespace,
		},
	}
	err := k8sClient.Delete(context.TODO(), &gw)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Kyma gateway")
	}

	if err == nil {
		ctrl.Log.Info("Successfully deleted Kyma gateway")
	}

	return nil
}
