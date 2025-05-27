package gateway

import (
	"context"
	_ "embed"
	"fmt"

	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/version"
)

const (
	kymaGatewayVirtualServiceName      = "istio-healthz"
	kymaGatewayVirtualServiceNamespace = "istio-system"
)

//go:embed virtual_service.yaml
var virtualServiceManifest []byte

func reconcileKymaGatewayVirtualService(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Virtual Service entry", "KymaGatewayEnabled", isEnabled, "name", kymaGatewayVirtualServiceName, "namespace", kymaGatewayVirtualServiceNamespace)

	if !isEnabled || apiGatewayCR.IsInDeletion() {
		return deleteVirtualService(ctx, k8sClient, kymaGatewayVirtualServiceName, kymaGatewayVirtualServiceNamespace)
	}

	return reconcileVirtualService(ctx, k8sClient, kymaGatewayVirtualServiceName, kymaGatewayVirtualServiceNamespace, domain)
}

func reconcileVirtualService(ctx context.Context, k8sClient client.Client, name, namespace, domain string) error {
	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = namespace
	templateValues["Domain"] = domain
	templateValues["Gateway"] = KymaGatewayFullName
	templateValues["Version"] = version.GetModuleVersion()

	return reconciliations.ApplyResource(ctx, k8sClient, virtualServiceManifest, templateValues)
}

func deleteVirtualService(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting Virtual Service if it exists", "name", name, "namespace", namespace)
	d := v1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &d)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete  Virtual Service %s/%s: %w", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Virtual Service as it wasn't present", "name", name, "namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Virtual Service", "name", name, "namespace", namespace)
	}

	return nil
}
