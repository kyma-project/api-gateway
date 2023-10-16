package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kymaGatewayVirtualServiceName           = "kyma-gateway"
	kymaGatewayVirtualServiceNamespace      = "kyma-system"
	kymaGatewayVirtualServicePort           = "15021"
	kymaGatewayVirtualServiceIngressGateway = "istio-ingressgateway.istio-system.svc.cluster.local"
)

//go:embed virtual_service.yaml
var virtualServiceManifest []byte

func reconcileKymaGatewayVirtualService(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	name := kymaGatewayVirtualServiceName
	namespace := kymaGatewayVirtualServiceNamespace
	port := kymaGatewayVirtualServicePort
	ingressGateway := kymaGatewayVirtualServiceIngressGateway

	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling Virtual Service entry", "KymaGatewayEnabled", isEnabled, "name", name, "namespace", namespace)

	if !isEnabled || apiGatewayCR.IsInDeletion() {
		return deleteVirtualService(ctx, k8sClient, name, namespace)
	}

	return reconcileVirtualService(ctx, k8sClient, name, namespace, domain, port, ingressGateway)
}

func reconcileVirtualService(ctx context.Context, k8sClient client.Client, name, namespace, domain, port, ingressGateway string) error {
	gateway := fmt.Sprintf("%s/%s", kymaGatewayName, kymaGatewayNamespace)

	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = namespace
	templateValues["Domain"] = domain
	templateValues["Gateway"] = gateway
	templateValues["Port"] = port
	templateValues["IngressGateway"] = ingressGateway

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
		return fmt.Errorf("failed to delete  Virtual Service %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of Virtual Service as it wasn't present", "name", name, "namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted Virtual Service", "name", name, "namespace", namespace)
	}

	return nil
}
