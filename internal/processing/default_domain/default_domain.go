package default_domain

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/thoas/go-funk"
	apiv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kymaGatewayNamespace = "kyma-system"
	kymaGatewayName      = "kyma-gateway"
	kymaGatewayProtocol  = "HTTPS"
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

func HandleDefaultDomainError(log logr.Logger, err error) (finishReconciliation bool) {
	if apierrs.IsNotFound(err) {
		log.Error(err, "Default domain wasn't found. APIRules will require full host")
		return false
	} else {
		log.Error(err, "Error getting default domain")
		return true
	}
}

func GetDomainFromKymaGateway(ctx context.Context, k8sClient client.Client) (string, error) {
	var gateway networkingv1beta1.Gateway
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: kymaGatewayNamespace, Name: kymaGatewayName}, &gateway)
	if err != nil {
		return "", err
	}

	httpsServers := funk.Filter(gateway.Spec.GetServers(), func(g *apiv1beta1.Server) bool {
		return g.Port != nil && strings.ToUpper(g.Port.Protocol) == kymaGatewayProtocol
	}).([]*apiv1beta1.Server)

	if len(httpsServers) != 1 {
		return "", fmt.Errorf("gateway must have a single https server definition, num=%d", len(httpsServers))
	}

	if len(httpsServers[0].Hosts) != 1 {
		return "", fmt.Errorf("gateway https server must have a single host definition, num=%d", len(httpsServers[0].Hosts))
	}

	if !strings.HasPrefix(httpsServers[0].Hosts[0], "*.") {
		return "", fmt.Errorf(`gateway https server host %s does not start with a prefix "*."`, httpsServers[0].Hosts[0])
	}

	return strings.TrimPrefix(httpsServers[0].Hosts[0], "*."), nil
}

func GetDomainFromGateway(ctx context.Context, k8sClient client.Client, gatewayName, gatewayNamespace string) (string, error) {
	var gateway networkingv1beta1.Gateway
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: gatewayNamespace, Name: gatewayName}, &gateway)
	if err != nil {
		return "", err
	}

	if !gatewayServersWithSameSingleHost(&gateway) {
		return "", errors.New("gateway must specify server(s) with the same single host")
	}

	if !strings.HasPrefix(gateway.Spec.Servers[0].Hosts[0], "*.") {
		return "", fmt.Errorf(`gateway https server host %s does not start with a prefix "*."`, gateway.Spec.Servers[0].Hosts[0])
	}

	return strings.TrimPrefix(gateway.Spec.Servers[0].Hosts[0], "*."), nil
}

func gatewayServersWithSameSingleHost(gateway *networkingv1beta1.Gateway) bool {
	host := ""
	for _, server := range gateway.Spec.Servers {
		if len(server.Hosts) > 1 {
			return false
		}
		if host == "" {
			host = server.Hosts[0]
		} else if host != server.Hosts[0] {
			return false
		}
	}
	return host != ""
}
