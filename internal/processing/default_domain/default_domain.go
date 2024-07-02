package default_domain

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/thoas/go-funk"
	apiv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	gatewayNamespace = "kyma-system"
	gatewayName      = "kyma-gateway"

	protocolHttps = "HTTPS"
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

func GetDefaultDomainFromKymaGateway(ctx context.Context, k8sClient client.Client) (string, error) {
	var gateway networkingv1beta1.Gateway
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: gatewayNamespace, Name: gatewayName}, &gateway)
	if err != nil {
		return "", err
	}

	httpsServers := funk.Filter(gateway.Spec.GetServers(), func(g *apiv1beta1.Server) bool {
		return g.Port != nil && strings.ToUpper(g.Port.Protocol) == protocolHttps
	}).([]*apiv1beta1.Server)

	if len(httpsServers) != 1 {
		return "", fmt.Errorf("could not get default domain, number of https servers was more than 1, num=%d", len(gateway.Spec.GetServers()))
	}

	if len(httpsServers[0].Hosts) != 1 {
		return "", fmt.Errorf("could not get default domain, number of hosts in HTTPS server was different than default of 1, num=%d", len(gateway.Spec.GetServers()[0].Hosts))
	}

	match, err := regexp.MatchString(`^\*\..+$`, httpsServers[0].Hosts[0])

	if err != nil {
		return "", err
	}

	if !match {
		return "", fmt.Errorf(`host %s isn't a host with wildcard prefix "*."`, httpsServers[0].Hosts[0])
	}

	return strings.TrimLeft(gateway.Spec.GetServers()[0].Hosts[0], "*."), nil
}
