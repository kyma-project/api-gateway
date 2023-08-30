package processing

import (
	"fmt"
	gatewayv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
)

var (
	//OwnerLabel .
	OwnerLabel = fmt.Sprintf("%s.%s", "apirule", v1beta1.GroupVersion.String())
	//OwnerLabelv1alpha1 .
	OwnerLabelv1alpha1 = fmt.Sprintf("%s.%s", "apirule", gatewayv1alpha1.GroupVersion.String())
)

func HasJwtRule(api *v1beta1.APIRule) bool {
	for _, rule := range api.Spec.Rules {
		if IsJwtSecured(rule) {
			return true
		}
	}
	return false
}

func IsJwtSecured(rule v1beta1.Rule) bool {
	for _, strat := range rule.AccessStrategies {
		if strat.Name == "jwt" {
			return true
		}
	}
	return false
}

func IsSecured(rule v1beta1.Rule) bool {
	if len(rule.Mutators) > 0 {
		return true
	}
	for _, strat := range rule.AccessStrategies {
		if strat.Name != "allow" {
			return true
		}
	}
	return false
}

func GetOwnerLabels(api *v1beta1.APIRule) map[string]string {
	labels := make(map[string]string)
	labels[OwnerLabelv1alpha1] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	return labels
}

func FilterDuplicatePaths(rules []v1beta1.Rule) []v1beta1.Rule {
	duplicates := make(map[string]bool)
	var filteredRules []v1beta1.Rule
	for _, rule := range rules {
		if _, exists := duplicates[rule.Path]; !exists {
			duplicates[rule.Path] = true
			filteredRules = append(filteredRules, rule)
		}
	}

	return filteredRules
}

func FilterAccessStrategies(accessStrategies []*v1beta1.Authenticator, includeAllow bool, includeOryOnly bool, includeJwt bool) []*v1beta1.Authenticator {
	filterFunc := func(auth *v1beta1.Authenticator) bool {
		return ((includeAllow && auth.Handler.Name == "allow") ||
			(includeOryOnly && (auth.Handler.Name == "noop" || auth.Handler.Name == "oauth2_introspection")) ||
			(includeJwt && auth.Handler.Name == "jwt"))
	}
	return FilterGeneric(accessStrategies, filterFunc)
}

func FilterGeneric[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
