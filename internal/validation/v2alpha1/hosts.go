package v2alpha1

import (
	"fmt"
	"strings"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

func validateHosts(parentAttributePath string, vsList networkingv1beta1.VirtualServiceList, gwList networkingv1beta1.GatewayList, apiRule *gatewayv2alpha1.APIRule) []validation.Failure {
	var failures []validation.Failure
	hostsAttributePath := parentAttributePath + ".hosts"

	hosts := apiRule.Spec.Hosts
	if len(hosts) == 0 {
		failures = append(failures, validation.Failure{
			AttributePath: hostsAttributePath,
			Message:       "No hosts defined",
		})
		return failures
	}

	for hostIndex, host := range hosts {
		gatewayDomain := ""
		if helpers.IsShortHostName(string(*host)) {
			gateway := findGateway(*apiRule.Spec.Gateway, gwList)
			if gateway == nil {
				hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
				failures = append(failures, validation.Failure{
					AttributePath: hostAttributePath,
					Message:       fmt.Sprintf(`Unable to find Gateway "%s"`, *apiRule.Spec.Gateway),
				})
			} else if !hasSingleHostDefinitionWithCorrectPrefix(gateway) {
				hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
				failures = append(failures, validation.Failure{
					AttributePath: hostAttributePath,
					Message:       "Lowercase RFC 1123 label (short host) is only supported as the APIRule host when selected Gateway has a single host definition matching *.<fqdn> format",
				})
			}
			gatewayDomain = getGatewayDomain(gateway)
		} else if !helpers.IsFqdnHostName(string(*host)) {
			hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
			failures = append(failures, validation.Failure{
				AttributePath: hostAttributePath,
				Message:       "Host must be a valid FQDN or short host name",
			})
		}
		for _, vs := range vsList.Items {
			hostWithDomain := default_domain.GetHostWithDomain(string(*host), gatewayDomain)
			if occupiesHost(vs, hostWithDomain) && !ownedBy(vs, apiRule) {
				hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
				failures = append(failures, validation.Failure{
					AttributePath: hostAttributePath,
					Message:       "Host is occupied by another Virtual Service",
				})
			}
		}
	}

	return failures
}

func getGatewayDomain(gateway *networkingv1beta1.Gateway) string {
	if gateway != nil {
		for _, server := range gateway.Spec.Servers {
			if len(server.Hosts) > 0 {
				return strings.TrimPrefix(server.Hosts[0], "*.")
			}
		}
	}
	return ""
}

func hasSingleHostDefinitionWithCorrectPrefix(gateway *networkingv1beta1.Gateway) bool {
	host := ""
	for _, server := range gateway.Spec.Servers {
		if len(server.Hosts) > 1 {
			return false
		}
		if !strings.HasPrefix(server.Hosts[0], "*.") {
			return false
		}
		if host == "" {
			host = server.Hosts[0]
		} else if host != server.Hosts[0] {
			return false
		}
	}
	return true
}

func findGateway(gatewayNamespacedName string, gwList networkingv1beta1.GatewayList) *networkingv1beta1.Gateway {
	for _, gateway := range gwList.Items {
		if gatewayNamespacedName == strings.Join([]string{gateway.Namespace, gateway.Name}, "/") {
			return gateway
		}
	}
	return nil
}

func occupiesHost(vs *networkingv1beta1.VirtualService, host string) bool {
	for _, h := range vs.Spec.Hosts {
		if h == host {
			return true
		}
	}
	return false
}

func ownedBy(vs *networkingv1beta1.VirtualService, apiRule *gatewayv2alpha1.APIRule) bool {
	ownerLabelKey, ownerLabelValue := getExpectedOwnerLabel(apiRule)
	vsLabels := vs.GetLabels()

	val, ok := vsLabels[ownerLabelKey]
	if ok {
		return val == ownerLabelValue
	} else {
		return false
	}
}

func getExpectedOwnerLabel(apiRule *gatewayv2alpha1.APIRule) (string, string) {
	// the label must use version v1beta1 for backward compatibility
	return fmt.Sprintf("%s.%s", "apirule", "gateway.kyma-project.io/v1beta1"),
		fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace)
}
