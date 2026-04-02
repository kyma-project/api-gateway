package gateway

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/version"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		return deleteDnsEntry(ctx, k8sClient, name, namespace)
	}

	istioIngressIps, serviceType, err := fetchIstioIngressGatewayIps(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to fetch Istio ingress gateway IP: %v", err)
	}

	return reconcileDnsEntry(ctx, k8sClient, name, namespace, domain, istioIngressIps, serviceType)
}

const ipStackAnnotation = "dns.gardener.cloud/ip-stack"

func reconcileDnsEntry(ctx context.Context, k8sClient client.Client, name, namespace, domain string, ingressGatewayIps []string, ipStackType string) error {
	templateValues := make(map[string]string)
	templateValues["Name"] = name
	templateValues["Namespace"] = namespace
	templateValues["Domain"] = domain
	templateValues["IngressGatewayServiceIps"] = fmt.Sprintf("[%s]", strings.Join(ingressGatewayIps, ","))
	templateValues["Version"] = version.GetModuleVersion()

	if ipStackType == ipStackTypeDualStack || ipStackType == ipStackTypeIPv6 {
		return reconciliations.ApplyResource(ctx, k8sClient, dnsEntryManifest, templateValues,
			reconciliations.WithApplyAnnotations(map[string]string{ipStackAnnotation: ipStackType}))
	}

	return reconciliations.ApplyResource(ctx, k8sClient, dnsEntryManifest, templateValues)
}

func deleteDnsEntry(ctx context.Context, k8sClient client.Client, name, namespace string) error {
	ctrl.Log.Info("Deleting DNSEntry if it exists", "name", name, "namespace", namespace)
	d := dnsv1alpha1.DNSEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := k8sClient.Delete(ctx, &d)

	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete DNSEntry %s/%s: %v", namespace, name, err)
	}

	if k8serrors.IsNotFound(err) {
		ctrl.Log.Info("Skipped deletion of DNSEntry as it wasn't present", "name", name, "namespace", namespace)
	} else {
		ctrl.Log.Info("Successfully deleted DNSEntry", "name", name, "namespace", namespace)
	}

	return nil
}

const (
	ipStackTypeIPv4      = "ipv4"
	ipStackTypeIPv6      = "ipv6"
	ipStackTypeDualStack = "dual-stack"
)

// fetchIstioIngressGatewayIps returns the external IPs and hostnames of the istio-ingressgateway service.
// The second return value indicates the type of the Service (IPv4, IPv6 or DualStack) based on IPFamilies field of the Service spec.
// In case the IPFamilies field is not set, it defaults to IPv4.
func fetchIstioIngressGatewayIps(ctx context.Context, k8sClient client.Client) ([]string, string, error) {
	istioIngressGatewayNamespaceName := types.NamespacedName{
		Name:      "istio-ingressgateway",
		Namespace: "istio-system",
	}

	svc := corev1.Service{}
	if err := k8sClient.Get(ctx, istioIngressGatewayNamespaceName, &svc); err != nil {
		return nil, "", err
	}

	stackType := ipStackTypeIPv4
	if len(svc.Spec.IPFamilies) == 2 {
		stackType = ipStackTypeDualStack
	} else if len(svc.Spec.IPFamilies) == 1 && svc.Spec.IPFamilies[0] == corev1.IPv6Protocol {
		stackType = ipStackTypeIPv6
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return nil, stackType, fmt.Errorf("no ingress exists for %s", istioIngressGatewayNamespaceName.String())
	}

	var targets []string
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		if ingress.IP != "" {
			targets = append(targets, ingress.IP)
		}
		if ingress.Hostname != "" {
			targets = append(targets, ingress.Hostname)
		}
	}

	ctrl.Log.Info("Found istio ingress gateway IP addresses", "targets", targets, "stackType", stackType)
	if len(targets) > 0 {
		return targets, stackType, nil
	}

	return nil, stackType, fmt.Errorf("no ingress targets found for %s", istioIngressGatewayNamespaceName.String())
}
