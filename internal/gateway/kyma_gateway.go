package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"

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

	templateValues := make(map[string]string)
	templateValues["Name"] = kymaGatewayName
	templateValues["Namespace"] = kymaGatewayNamespace
	templateValues["Domain"] = domain
	templateValues["CertificateSecretName"] = kymaGatewayCertSecretName

	resource, err := createUnstructuredResource(kymaGatewayManifest, templateValues)
	if err != nil {
		return err
	}

	if !isEnabled || apiGatewayCR.IsInGracefulDeletion() {
		return deleteKymaGateway(ctx, k8sClient, resource)
	}

	return createOrUpdateResource(ctx, k8sClient, resource)
}

func isKymaGatewayEnabled(cr v1alpha1.APIGateway) bool {
	return cr.Spec.EnableKymaGateway != nil && *cr.Spec.EnableKymaGateway == true
}

func deleteKymaGateway(ctx context.Context, k8sClient client.Client, kymaGateway unstructured.Unstructured) error {
	ctrl.Log.Info("Deleting Kyma gateway if it exists")
	err := k8sClient.Delete(ctx, &kymaGateway)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Kyma gateway: %v", err)
	}

	ctrl.Log.Info("Successfully deleted Kyma gateway")
	return nil
}
