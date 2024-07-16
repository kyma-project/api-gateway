package v2alpha1

import (
	"fmt"
	"regexp"
	"strings"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

const (
	fqdnMaxLength        = 253
	fqdnMinSegments      = 2
	domainLabelMaxLength = 63
)

var (
	dnsLabelRegexp = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
)

func validateHosts(parentAttributePath string, vsList networkingv1beta1.VirtualServiceList, apiRule *gatewayv2alpha1.APIRule) []validation.Failure {
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
		if !isFQDN(string(*host)) {
			hostAttributePath := fmt.Sprintf("%s[%d]", hostsAttributePath, hostIndex)
			failures = append(failures, validation.Failure{
				AttributePath: hostAttributePath,
				Message:       "Host is not fully qualified domain name",
			})
		}
	}

	for _, vs := range vsList.Items {
		for hostIndex, host := range hosts {
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

func isFQDN(host string) bool {
	if len(host) > fqdnMaxLength {
		return false
	}

	labels := strings.Split(host, ".")
	if len(labels) < fqdnMinSegments {
		return false
	}

	for _, domainLabel := range labels {
		if len(domainLabel) == 0 || len(domainLabel) > domainLabelMaxLength {
			return false
		}
		if !dnsLabelRegexp.MatchString(domainLabel) {
			return false
		}
	}

	return true
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
