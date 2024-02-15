package istio

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
)

// NewAccessRuleProcessor returns a AccessRuleProcessor with the desired state handling specific for the Ory handler.
func NewAccessRuleProcessor(config processing.ReconciliationConfig) processors.AccessRuleProcessor {
	return processors.AccessRuleProcessor{
		Creator: accessRuleCreator{
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type accessRuleCreator struct {
	additionalLabels  map[string]string
	defaultDomainName string
}

// Create returns a map of rules using the configuration of the APIRule. The key of the map is a unique combination of
// the match URL and methods of the rule.
func (r accessRuleCreator) Create(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	pathDuplicates := processors.HasPathDuplicates(api.Spec.Rules)
	accessRules := make(map[string]*rulev1alpha1.Rule)
	for _, rule := range api.Spec.Rules {
		filteredAS := filterAccessStrategies(rule.AccessStrategies)
		if len(filteredAS) > 0 && processing.IsSecuredByOathkeeper(rule) {
			ar := processors.GenerateAccessRule(api, rule, filteredAS, r.additionalLabels, r.defaultDomainName)
			accessRules[processors.SetAccessRuleKey(pathDuplicates, *ar)] = ar
		}
	}
	return accessRules
}

func filterAccessStrategies(accessStrategies []*gatewayv1beta1.Authenticator) []*gatewayv1beta1.Authenticator {
	filterFunc := func(auth *gatewayv1beta1.Authenticator) bool {
		return auth.Handler.Name == gatewayv1beta1.AccessStrategyNoop || auth.Handler.Name == gatewayv1beta1.AccessStrategyOauth2Introspection
	}

	return filterGeneric(accessStrategies, filterFunc)
}

func filterGeneric[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
