package helpers

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

func FindServiceNamespace(api *gatewayv1beta1.APIRule, rule *gatewayv1beta1.Rule) string {
	// Fallback direction for the upstream service namespace: Rule.Service > Spec.Service > APIRule
	if rule != nil && rule.Service != nil && rule.Service.Namespace != nil {
		return *rule.Service.Namespace
	}
	if api.Spec.Service != nil && api.Spec.Service.Namespace != nil {
		return *api.Spec.Service.Namespace
	}
	return api.Namespace
}
