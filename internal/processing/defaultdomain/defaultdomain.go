package defaultdomain

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

func GetHostWithDomain(host, domainName string) string {
	if !HostIncludesDomain(host) && domainName != "" {
		return BuildHostWithDomain(host, domainName)
	}
	return host
}

func HostIncludesDomain(host string) bool {
	return strings.Contains(host, ".")
}

func BuildHostWithDomain(host, domainName string) string {
	return fmt.Sprintf("%s.%s", host, domainName)
}

func GetHostLocalDomain(host string, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", host, namespace)
}

func HandleDefaultDomainError(log logr.Logger, err error) (finishReconciliation bool) {
	if apierrs.IsNotFound(err) {
		log.Error(err, "Default domain wasn't found. APIRules will require full host")
		return false
	}
	log.Error(err, "Error getting default domain")
	return true
}

func GetDomainFromKymaGateway(ctx context.Context, k8sClient client.Client) (string, error) {
	var gateway networkingv1beta1.Gateway
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: kymaGatewayNamespace, Name: kymaGatewayName}, &gateway)
	if err != nil {
		return "", err
	}

	httpsServers := funk.Filter(gateway.Spec.GetServers(), func(g *apiv1beta1.Server) bool {
		return g.Port != nil && strings.ToUpper(g.Port.GetProtocol()) == kymaGatewayProtocol
	}).([]*apiv1beta1.Server)

	if len(httpsServers) != 1 {
		return "", fmt.Errorf("gateway must have a single https server definition, num=%d", len(httpsServers))
	}

	if len(httpsServers[0].Hosts) != 1 {
		return "", fmt.Errorf("gateway https server must have a single host definition, num=%d", len(httpsServers[0].Hosts))
	}

	if !strings.HasPrefix(httpsServers[0].Hosts[0], "*.") {
		return "", fmt.Errorf(`gateway https server host "%s" does not start with the prefix "*."`, httpsServers[0].Hosts[0])
	}

	domain := strings.TrimPrefix(httpsServers[0].Hosts[0], "*.")
	if domain == "" {
		return "", fmt.Errorf(`gateway https server host "%s" does not define domain after the prefix "*."`, httpsServers[0].Hosts[0])
	}

	return domain, nil
}

func GetDomainFromGateway(ctx context.Context, k8sClient client.Client, gatewayName, gatewayNamespace string) (string, error) {
	var gateway networkingv1beta1.Gateway
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: gatewayNamespace, Name: gatewayName}, &gateway)
	if err != nil {
		return "", err
	}

	serverHost := ""
	for _, server := range gateway.Spec.GetServers() {
		switch len(server.GetHosts()) {
		case 0: // ignored
		case 1:
			if serverHost == "" {
				serverHost = server.GetHosts()[0]
			} else if serverHost != server.GetHosts()[0] {
				return "", errors.New("gateway must have server definition(s) with the same host")
			}
		default:
			return "", errors.New("gateway must have server definition(s) with a single host")
		}
	}

	if !strings.HasPrefix(serverHost, "*.") {
		return "", fmt.Errorf(`gateway server host "%s" does not start with the prefix "*."`, serverHost)
	}

	domain := strings.TrimPrefix(serverHost, "*.")
	if domain == "" {
		return "", fmt.Errorf(`gateway server host "%s" does not define domain after the prefix "*."`, serverHost)
	}

	return domain, nil
}
