package v2alpha1

import (
	"fmt"
	"strings"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/helpers"
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
		if !helpers.IsHostFqdn(string(*host)) {
			if helpers.IsHostShortName(string(*host)) { // short name
				gateway := findGateway(*apiRule.Spec.Gateway, gwList)
				if gateway == nil {
					hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
					failures = append(failures, validation.Failure{
						AttributePath: hostAttributePath,
						Message:       fmt.Sprintf("Unable to find Gateway %s", *apiRule.Spec.Gateway),
					})
				} else if hasMultipleHostDefinitions(gateway) {
					hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
					failures = append(failures, validation.Failure{
						AttributePath: hostAttributePath,
						Message:       "Short host only supported when Gateway has single host definition matching *.<fqdn> format",
					})
				}
			} else {
				hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
				failures = append(failures, validation.Failure{
					AttributePath: hostAttributePath,
					Message:       "Host must be a valid FQDN or short name",
				})
			}
		}
		for _, vs := range vsList.Items {
			if occupiesHost(vs, string(*host)) && !ownedBy(vs, apiRule) {
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

func hasMultipleHostDefinitions(gateway *networkingv1beta1.Gateway) bool {
	host := ""
	for _, server := range gateway.Spec.Servers {
		if len(server.Hosts) > 1 {
			return true
		}
		if !strings.HasPrefix(server.Hosts[0], "*.") {
			return true
		}
		if host == "" {
			host = server.Hosts[0]
		} else if host != server.Hosts[0] {
			return true
		}
	}
	return false
}

func findGateway(name string, gwList networkingv1beta1.GatewayList) *networkingv1beta1.Gateway {
	for _, gateway := range gwList.Items {
		if gateway.Name == name {
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
	return fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String()),
		fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace)
}
