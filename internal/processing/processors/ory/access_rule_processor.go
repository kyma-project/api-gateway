package ory

import (
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
)

// NewAccessRuleProcessor returns a AccessRuleProcessor with the desired state handling specific for the Ory handler.
func NewAccessRuleProcessor(config processing.ReconciliationConfig, apiRule *gatewayv1beta1.APIRule) processors.AccessRuleProcessor {
	return processors.AccessRuleProcessor{
		ApiRule: apiRule,
		Creator: accessRuleCreator{
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type accessRuleCreator struct {
	defaultDomainName string
}

// Create returns a map of rules using the configuration of the APIRule. The key of the map is a unique combination of
// the match URL and methods of the rule.
func (r accessRuleCreator) Create(api *gatewayv1beta1.APIRule) map[string]*rulev1alpha1.Rule {
	pathDuplicates := processors.HasPathDuplicates(api.Spec.Rules)
	accessRules := make(map[string]*rulev1alpha1.Rule)
	for _, rule := range api.Spec.Rules {
		if processing.IsSecuredByOathkeeper(rule) {
			ar := processors.GenerateAccessRule(api, rule, rule.AccessStrategies, r.defaultDomainName)
			accessRules[processors.SetAccessRuleKey(pathDuplicates, *ar)] = ar
		}
	}
	return accessRules
}
