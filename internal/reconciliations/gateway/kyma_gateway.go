package gateway

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/version"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	KymaGatewayName      = "kyma-gateway"
	KymaGatewayNamespace = "kyma-system"
	KymaGatewayFullName  = "kyma-system/kyma-gateway"
)

//go:embed kyma_gateway.yaml
var kymaGatewayManifest []byte

func reconcileKymaGateway(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Kyma gateway", "KymaGatewayEnabled", isEnabled)

	templateValues := make(map[string]string)
	templateValues["Name"] = KymaGatewayName
	templateValues["Namespace"] = KymaGatewayNamespace
	templateValues["Domain"] = domain
	templateValues["CertificateSecretName"] = kymaGatewayCertSecretName
	templateValues["Version"] = version.GetModuleVersion()

	resource, err := reconciliations.CreateUnstructuredResource(kymaGatewayManifest, templateValues)
	if err != nil {
		return err
	}

	if !isEnabled || apiGatewayCR.IsInDeletion() {
		return deleteKymaGateway(ctx, k8sClient, resource)
	}

	return reconciliations.CreateOrUpdateResource(ctx, k8sClient, resource)
}

func isKymaGatewayEnabled(cr v1alpha1.APIGateway) bool {
	return cr.Spec.EnableKymaGateway != nil && *cr.Spec.EnableKymaGateway
}

func deleteKymaGateway(ctx context.Context, k8sClient client.Client, kymaGateway unstructured.Unstructured) error {
	ctrl.Log.Info("Deleting Kyma gateway")
	err := k8sClient.Delete(ctx, &kymaGateway)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Kyma gateway: %v", err)
	}

	if errors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Kyma gateway as it wasn't present")
	} else {
		ctrl.Log.Info("Successfully deleted Kyma gateway")
	}

	return nil
}
