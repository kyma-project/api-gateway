package processing

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isSecured(rule gatewayv1beta1.Rule) bool {
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

func hasPathDuplicates(rules []gatewayv1beta1.Rule) bool {
	duplicates := map[string]bool{}
	for _, rule := range rules {
		if duplicates[rule.Path] {
			return true
		}
		duplicates[rule.Path] = true
	}

	return false
}

func filterDuplicatePaths(rules []gatewayv1beta1.Rule) []gatewayv1beta1.Rule {
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

func setAccessRuleKey(hasPathDuplicates bool, rule rulev1alpha1.Rule) string {
	if hasPathDuplicates {
		return fmt.Sprintf("%s:%s", rule.Spec.Match.URL, rule.Spec.Match.Methods)
	}

	return rule.Spec.Match.URL
}

func setAuthorizationPolicyKey(hasPathDuplicates bool, ap *istiosecurityv1beta1.AuthorizationPolicy) string {
	if hasPathDuplicates {
		return fmt.Sprintf("%s:%s", ap.Spec.Rules[0].To[0].Operation.Paths, ap.Spec.Rules[0].To[0].Operation.Methods)
	}

	return ap.Spec.Rules[0].To[0].Operation.Paths[0]
}

func setRequestAuthenticationKey(ra *istiosecurityv1beta1.RequestAuthentication) string {
	return fmt.Sprintf("%s", ra.Spec.JwtRules)
}

func generateOwnerRef(api *gatewayv1beta1.APIRule) k8sMeta.OwnerReference {
	return *builders.OwnerReference().
		Name(api.ObjectMeta.Name).
		APIVersion(api.TypeMeta.APIVersion).
		Kind(api.TypeMeta.Kind).
		UID(api.ObjectMeta.UID).
		Controller(true).
		Get()
}
