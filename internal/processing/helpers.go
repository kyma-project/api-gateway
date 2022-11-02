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

func isJwtSecured(rule gatewayv1beta1.Rule) bool {
	for _, strat := range rule.AccessStrategies {
		if strat.Name == "jwt" {
			return true
		}
	}
	return false
}

func checkPathDuplicates(rules []gatewayv1beta1.Rule) bool {
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

func getAuthorizationPolicyKey(hasPathDuplicates bool, ap *istiosecurityv1beta1.AuthorizationPolicy) string {
	key := ""
	if ap.Spec.Rules != nil && len(ap.Spec.Rules) > 0 && ap.Spec.Rules[0].To != nil && len(ap.Spec.Rules[0].To) > 0 {
		if hasPathDuplicates {
			key = fmt.Sprintf("%s:%s",
				sliceToString(ap.Spec.Rules[0].To[0].Operation.Paths),
				sliceToString(ap.Spec.Rules[0].To[0].Operation.Methods))
		} else {
			key = sliceToString(ap.Spec.Rules[0].To[0].Operation.Paths)
		}
	}

	return key
}

func getRequestAuthenticationKey(ra *istiosecurityv1beta1.RequestAuthentication) string {
	key := ""
	for _, k := range ra.Spec.JwtRules {
		key += fmt.Sprintf("%s:%s", k.Issuer, k.JwksUri)
	}
	return key
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

func sliceToString(ss []string) (s string) {
	for _, el := range ss {
		s += el
	}
	return
}
