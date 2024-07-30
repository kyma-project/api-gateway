package processing

import (
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
)

var (
	//OwnerLabel .
	OwnerLabel = fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())
)

func RequiresAuthorizationPolicies(api *gatewayv1beta1.APIRule) bool {
	for _, rule := range api.Spec.Rules {
		if IsJwtSecured(rule) || rule.ContainsAccessStrategy(gatewayv1beta1.AccessStrategyNoAuth) {
			return true
		}
	}
	return false
}

func IsJwtSecured(rule gatewayv1beta1.Rule) bool {
	for _, strat := range rule.AccessStrategies {
		if strat.Name == "jwt" {
			return true
		}
	}
	return false
}

// IsSecuredByOathkeeper checks whether the rule contains an access strategy that should lead to the creation of an Oathkeeper rule.
func IsSecuredByOathkeeper(rule gatewayv1beta1.Rule) bool {
	for _, strat := range rule.AccessStrategies {
		if strat.Name != gatewayv1beta1.AccessStrategyAllow && strat.Name != gatewayv1beta1.AccessStrategyNoAuth {
			return true
		}
	}
	return false
}

func GetOwnerLabels(api *gatewayv1beta1.APIRule) map[string]string {
	labels := make(map[string]string)
	labels[OwnerLabel] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	return labels
}

func FilterDuplicatePaths(rules []gatewayv1beta1.Rule) []gatewayv1beta1.Rule {
	duplicates := make(map[string]bool)
	var filteredRules []gatewayv1beta1.Rule
	for _, rule := range rules {
		if _, exists := duplicates[rule.Path]; !exists {
			duplicates[rule.Path] = true
			filteredRules = append(filteredRules, rule)
		}
	}

	return filteredRules
}
