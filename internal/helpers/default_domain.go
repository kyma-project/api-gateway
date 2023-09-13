package helpers

import (
	"context"
	"fmt"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	gatewayNamespace = "kyma-system"
	gatewayName      = "kyma-gateway"
)

func GetHostWithDomain(host, defaultDomainName string) string {
	if !HostIncludesDomain(host) {
		return GetHostWithDefaultDomain(host, defaultDomainName)
	}
	return host
}

func HostIncludesDomain(host string) bool {
	return strings.Contains(host, ".")
}

func GetHostWithDefaultDomain(host, defaultDomainName string) string {
	return fmt.Sprintf("%s.%s", host, defaultDomainName)
}

func GetHostLocalDomain(host string, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", host, namespace)
}

func GetDefaultDomainFromKymaGateway(ctx context.Context, k8sClient client.Client) (string, error) {
	var gateway networkingv1beta1.Gateway
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: gatewayNamespace, Name: gatewayName}, &gateway)
	if err != nil {
		return "", err
	}

	if len(gateway.Spec.GetServers()) != 2 {
		return "", fmt.Errorf("could not get default domain, number of servers was different than default of 2, num=%d", len(gateway.Spec.GetServers()))
	}

	if len(gateway.Spec.GetServers()[0].Hosts) != 1 {
		return "", fmt.Errorf("could not get default domain, number of hosts in HTTPS server was different than default of 1, num=%d", len(gateway.Spec.GetServers()[0].Hosts))
	}

	return strings.TrimLeft(gateway.Spec.GetServers()[0].Hosts[0], "*."), nil
}
