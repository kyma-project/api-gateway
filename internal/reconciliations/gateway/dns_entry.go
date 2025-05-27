package gateway

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

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/version"
)

const (
	kymaGatewayDnsEntryName      = "kyma-gateway"
	kymaGatewayDnsEntryNamespace = "kyma-system"
)

//go:embed dns_entry.yaml
var dnsEntryManifest []byte

func reconcileKymaGatewayDnsEntry(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway, domain string) error {
	name := kymaGatewayDnsEntryName
	namespace := kymaGatewayDnsEntryNamespace

	isEnabled := isKymaGatewayEnabled(apiGatewayCR)
	ctrl.Log.Info("Reconciling DNS entry", "KymaGatewayEnabled", isEnabled, "name", name, "namespace", namespace)

	if !isEnabled || apiGatewayCR.IsInDeletion() {
		return deleteDNSEntry(ctx, k8sClient, name, namespace)
	}

	istioIngressIp, err := fetchIstioIngressGatewayIP(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to fetch Istio ingress gateway IP: %w", err)
	}

	return reconcileDNSEntry(ctx, k8sClient, name, namespace, domain, istioIngressIp)
}

func reconcileDNSEntry(ctx context.Context, k8sClient client.Client, name, namespace, domain, ingressGatewayIP string) error {
	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = namespace
	templateValues["Domain"] = domain
	templateValues["IngressGatewayServiceIp"] = ingressGatewayIP
	templateValues["Version"] = version.GetModuleVersion()

	return reconciliations.ApplyResource(ctx, k8sClient, dnsEntryManifest, templateValues)
}

func deleteDNSEntry(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting DNSEntry if it exists", "name", name, "namespace", namespace)
	d := dnsv1alpha1.DNSEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &d)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete DNSEntry %s/%s: %w", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of DNSEntry as it wasn't present", "name", name, "namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted DNSEntry", "name", name, "namespace", namespace)
	}

	return nil
}

func fetchIstioIngressGatewayIP(ctx context.Context, k8sClient client.Client) (string, error) {
	istioIngressGatewayNamespaceName := types.NamespacedName{
		Name:      "istio-ingressgateway",
		Namespace: "istio-system",
	}

	svc := corev1.Service{}
	if err := k8sClient.Get(ctx, istioIngressGatewayNamespaceName, &svc); err != nil {
		return "", err
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", fmt.Errorf("no ingress exists for %s", istioIngressGatewayNamespaceName.String())
	}

	if svc.Status.LoadBalancer.Ingress[0].IP != "" {
		return svc.Status.LoadBalancer.Ingress[0].IP, nil
	}

	ctrl.Log.Info("Load balancer ingress IP is not set, trying to get hostname", "service", svc.Name, "namespace", svc.Namespace)

	if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
		return svc.Status.LoadBalancer.Ingress[0].Hostname, nil
	}

	ctrl.Log.Info("Load balancer ingress hostname and IP is not set", "service", svc.Name, "namespace", svc.Namespace)
	return "", fmt.Errorf("no ingress ip set for %s", istioIngressGatewayNamespaceName.String())
}
