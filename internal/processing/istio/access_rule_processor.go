package istio

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
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
		filteredAS := processing.FilterAccessStrategies(rule.AccessStrategies, false, true, false)
		if len(filteredAS) > 0 && processing.IsSecured(rule) {
			ar := processors.GenerateAccessRule(api, rule, filteredAS, r.additionalLabels, r.defaultDomainName)
			accessRules[processors.SetAccessRuleKey(pathDuplicates, *ar)] = ar
		}
	}
	return accessRules
}
