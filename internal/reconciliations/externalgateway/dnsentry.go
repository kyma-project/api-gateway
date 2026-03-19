package externalgateway

import (
	"context"
	_ "embed"
	"fmt"

	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
)

//go:embed dnsentry.yaml
var dnsEntryManifest []byte

// ReconcileDNSEntry creates or updates the Gardener DNSEntry for the external gateway
func ReconcileDNSEntry(ctx context.Context, k8sClient client.Client, external *externalv1alpha1.ExternalGateway, internalDomain string) error {
	dnsName := fmt.Sprintf("%s-dns", external.GatewayName())
	wildcardDomain := fmt.Sprintf("*.%s", internalDomain)

	ctrl.Log.Info("Reconciling DNSEntry", "name", dnsName, "namespace", istioSystemNamespace, "domain", wildcardDomain)

	// Fetch Istio Ingress Gateway IP
	istioIngressIp, err := fetchIstioIngressGatewayIp(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to fetch Istio ingress gateway IP: %w", err)
	}

	templateValues := map[string]string{
		"Name":                     dnsName,
		"Namespace":                istioSystemNamespace,
		"Domain":                   internalDomain,
		"IngressGatewayServiceIp":  istioIngressIp,
		"ExternalGatewayName":      external.Name,
		"ExternalGatewayNamespace": external.Namespace,
	}

	return reconciliations.ApplyResource(ctx, k8sClient, dnsEntryManifest, templateValues)
}

// DeleteDNSEntry deletes the DNSEntry resource
func DeleteDNSEntry(ctx context.Context, k8sClient client.Client, gatewayName string) error {
	dnsName := fmt.Sprintf("%s-dns", gatewayName)

	ctrl.Log.Info("Deleting DNSEntry if it exists", "name", dnsName, "namespace", istioSystemNamespace)

	dnsEntry := &dnsv1alpha1.DNSEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dnsName,
			Namespace: istioSystemNamespace,
		},
	}

	err := k8sClient.Delete(ctx, dnsEntry)
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete DNSEntry %s/%s: %w", istioSystemNamespace, dnsName, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of DNSEntry as it wasn't present", "name", dnsName)
	} else {
		ctrl.Log.Info("Successfully deleted DNSEntry", "name", dnsName)
	}

	return nil
}

// fetchIstioIngressGatewayIp retrieves the LoadBalancer IP of the Istio Ingress Gateway service
func fetchIstioIngressGatewayIp(ctx context.Context, k8sClient client.Client) (string, error) {
	istioIngressGatewayNamespaceName := types.NamespacedName{
		Name:      "istio-ingressgateway",
		Namespace: istioSystemNamespace,
	}

	var svc corev1.Service
	if err := k8sClient.Get(ctx, istioIngressGatewayNamespaceName, &svc); err != nil {
		return "", fmt.Errorf("failed to get istio-ingressgateway service: %w", err)
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("no ingress exists for %s", istioIngressGatewayNamespaceName.String())
	}

	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.IP != "" {
			ctrl.Log.Info("Found Istio Ingress Gateway IP", "ip", ingress.IP)
			return ingress.IP, nil
		}
		if ingress.Hostname != "" {
			ctrl.Log.Info("Found Istio Ingress Gateway hostname", "hostname", ingress.Hostname)
			return ingress.Hostname, nil
		}
	}

	return "", fmt.Errorf("no ingress ip set for %s", istioIngressGatewayNamespaceName.String())
}
