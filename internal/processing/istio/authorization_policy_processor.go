package istio

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/processing/processors"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// AuthorizationPolicyProcessor is the generic processor that handles the Istio Authorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	Creator AuthorizationPolicyCreator
}

// AuthorizationPolicyCreator provides the creation of AuthorizationPolicies using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type AuthorizationPolicyCreator interface {
	Create(api *gatewayv1beta1.APIRule) []*securityv1beta1.AuthorizationPolicy
}

// NewAuthorizationPolicyProcessor returns a AuthorizationPolicyProcessor with the desired state handling specific for the Istio handler.
func NewAuthorizationPolicyProcessor(config processing.ReconciliationConfig) processors.AuthorizationPolicyProcessor {
	return processors.AuthorizationPolicyProcessor{
		Creator: authorizationPolicyCreator{
			additionalLabels: config.AdditionalLabels,
		},
	}
}

type authorizationPolicyCreator struct {
	additionalLabels map[string]string
}

// Create returns the Authorization Policy using the configuration of the APIRule.
func (r authorizationPolicyCreator) Create(api *gatewayv1beta1.APIRule) []*securityv1beta1.AuthorizationPolicy {
	var authorizationPolicies []*securityv1beta1.AuthorizationPolicy
	hasJwtRule := processing.HasJwtRule(api)
	if hasJwtRule {
		for _, rule := range api.Spec.Rules {
			aps := generateAuthorizationPolicies(api, rule, r.additionalLabels)
			authorizationPolicies = append(authorizationPolicies, aps.Items...)
		}
	}
	return authorizationPolicies
}

func generateAuthorizationPolicies(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string) *securityv1beta1.AuthorizationPolicyList {
	authorizationPolicyList := securityv1beta1.AuthorizationPolicyList{}
	ruleAuthorizations := rule.GetAuthorizations()

	if len(ruleAuthorizations) == 0 {
		ap := generateAuthorizationPolicy(api, rule, additionalLabels, &gatewayv1beta1.Authorization{})
		authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
	} else {
		for _, authorization := range ruleAuthorizations {
			ap := generateAuthorizationPolicy(api, rule, additionalLabels, authorization)
			authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
		}
	}

	return &authorizationPolicyList
}

func generateAuthorizationPolicy(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string, authorization *gatewayv1beta1.Authorization) *securityv1beta1.AuthorizationPolicy {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := processing.GenerateOwnerRef(api)

	apBuilder := builders.AuthorizationPolicyBuilder().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AuthorizationPolicySpecBuilder().From(generateAuthorizationPolicySpec(api, rule, authorization))).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(processing.OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		apBuilder.Label(k, v)
	}

	return apBuilder.Get()
}

func generateAuthorizationPolicySpec(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, authorization *gatewayv1beta1.Authorization) *v1beta1.AuthorizationPolicy {
	var serviceName string
	if rule.Service != nil {
		serviceName = *rule.Service.Name
	} else {
		serviceName = *api.Spec.Service.Name
	}

	authorizationPolicySpec := builders.AuthorizationPolicySpecBuilder().
		Selector(builders.SelectorBuilder().MatchLabels(processors.AuthorizationPolicyAppSelectorLabel, serviceName))

	defaultScopeKeys := []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"}
	for _, scope := range defaultScopeKeys {
		generatedRule := generateAuthorizationPolicySpecRule(rule, scope, authorization)
		authorizationPolicySpec.Rule(generatedRule)
		// if requiredScopes are empty, only one rule is needed
		if generatedRule.Get().When == nil || len(generatedRule.Get().When) == 0 {
			break
		}
	}

	return authorizationPolicySpec.Get()
}

func generateAuthorizationPolicySpecRule(rule gatewayv1beta1.Rule, scope string, authorization *gatewayv1beta1.Authorization) *builders.Rule {
	ruleBuilder := builders.RuleBuilder().RuleTo(builders.RuleToBuilder().
		Operation(builders.OperationBuilder().Methods(rule.Methods).Path(rule.Path)))

	if processing.IsJwtSecured(rule) {
		ruleBuilder.RuleFrom(builders.RuleFromBuilder().Source())
	} else if processing.IsSecured(rule) {
		ruleBuilder.RuleFrom(builders.RuleFromBuilder().OathkeeperProxySource())
	} else {
		ruleBuilder.RuleFrom(builders.RuleFromBuilder().IngressGatewaySource())
	}

	ruleBuilder.RuleCondition(builders.RuleConditionBuilder().From(scope, authorization))

	return ruleBuilder
}
