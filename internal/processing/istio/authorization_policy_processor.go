package istio

import (
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	hashablestate "github.com/kyma-project/api-gateway/internal/processing/hashstate"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

const (
	audienceKey string = "request.auth.claims[aud]"
)

var (
	defaultScopeKeys = []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"}
)

// AuthorizationPolicyProcessor is the generic processor that handles the Istio JwtAuthorization Policies in the reconciliation of API Rule.
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

// Create returns the JwtAuthorization Policy using the configuration of the APIRule.
func (r authorizationPolicyCreator) Create(api *gatewayv1beta1.APIRule) (hashablestate.Desired, error) {
	state := hashablestate.NewDesired()
	hasJwtRule := processing.HasJwtRule(api)
	if hasJwtRule {
		for _, rule := range api.Spec.Rules {
			aps, err := generateAuthorizationPolicies(api, rule, r.additionalLabels)
			if err != nil {
				return state, nil
			}

			for _, ap := range aps.Items {
				err := state.Add(ap)

				if err != nil {
					return state, err
				}
			}
		}
	}
	return state, nil
}

func generateAuthorizationPolicies(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string) (*securityv1beta1.AuthorizationPolicyList, error) {
	authorizationPolicyList := securityv1beta1.AuthorizationPolicyList{}
	ruleAuthorizations := rule.GetJwtIstioAuthorizations()

	if len(ruleAuthorizations) == 0 {
		ap := generateAuthorizationPolicy(api, rule, additionalLabels, &gatewayv1beta1.JwtAuthorization{})

		// If there is no other authorization we can safely assume that the index of this authorization in the array
		// in the yaml is 0.
		err := hashablestate.AddHashingLabels(ap, 0)
		if err != nil {
			return &authorizationPolicyList, err
		}

		authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
	} else {
		for indexInYaml, authorization := range ruleAuthorizations {
			ap := generateAuthorizationPolicy(api, rule, additionalLabels, authorization)

			err := hashablestate.AddHashingLabels(ap, indexInYaml)
			if err != nil {
				return &authorizationPolicyList, err
			}

			authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
		}
	}

	return &authorizationPolicyList, nil
}

func generateAuthorizationPolicy(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string, authorization *gatewayv1beta1.JwtAuthorization) *securityv1beta1.AuthorizationPolicy {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := helpers.FindServiceNamespace(api, &rule)

	apBuilder := builders.NewAuthorizationPolicyBuilder().
		WithGenerateName(namePrefix).
		WithNamespace(namespace).
		WithSpec(builders.NewAuthorizationPolicySpecBuilder().FromAP(generateAuthorizationPolicySpec(api, rule, authorization)).Get()).
		WithLabel(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		WithLabel(processing.OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	// TODO: Why do we only need to add it in this case? It would be more reliable to always add the label. Of course if a new authorization is added in front of this would be evaluated as a change.
	//if len(index) > 0 {
	//	apBuilder.WithLabel(processing.IndexLabelName, strconv.Itoa(index[0]))
	//}

	for k, v := range additionalLabels {
		apBuilder.WithLabel(k, v)
	}

	return apBuilder.Get()
}

func generateAuthorizationPolicySpec(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, authorization *gatewayv1beta1.JwtAuthorization) *v1beta1.AuthorizationPolicy {
	var serviceName string
	if rule.Service != nil {
		serviceName = *rule.Service.Name
	} else {
		serviceName = *api.Spec.Service.Name
	}

	authorizationPolicySpecBuilder := builders.NewAuthorizationPolicySpecBuilder().
		WithSelector(builders.NewSelectorBuilder().WithMatchLabels(processors.AuthorizationPolicyAppSelectorLabel, serviceName).Get())

	// If RequiredScopes are configured, we need to generate a seperate Rule for each scopeKey in defaultScopeKeys
	if len(authorization.RequiredScopes) > 0 {
		for _, scopeKey := range defaultScopeKeys {
			ruleBuilder := baseRuleBuilder(rule)
			for _, scope := range authorization.RequiredScopes {
				ruleBuilder.WithWhenCondition(
					builders.NewConditionBuilder().WithKey(scopeKey).WithValues([]string{scope}).Get())
			}

			for _, aud := range authorization.Audiences {
				ruleBuilder.WithWhenCondition(
					builders.NewConditionBuilder().WithKey(audienceKey).WithValues([]string{aud}).Get())
			}

			authorizationPolicySpecBuilder.WithRule(ruleBuilder.Get())
		}
	} else { // Only one AP rule should be generated for other scenarios
		ruleBuilder := baseRuleBuilder(rule)
		for _, aud := range authorization.Audiences {
			ruleBuilder.WithWhenCondition(
				builders.NewConditionBuilder().WithKey(audienceKey).WithValues([]string{aud}).Get())
		}
		authorizationPolicySpecBuilder.WithRule(ruleBuilder.Get())
	}

	return authorizationPolicySpecBuilder.Get()
}

func withTo(b *builders.RuleBuilder, rule gatewayv1beta1.Rule) *builders.RuleBuilder {
	return b.WithTo(
		builders.NewToBuilder().
			WithOperation(builders.NewOperationBuilder().
				WithMethods(rule.Methods).WithPath(rule.Path).Get()).
			Get())
}

func withFrom(b *builders.RuleBuilder, rule gatewayv1beta1.Rule) *builders.RuleBuilder {
	if processing.IsJwtSecured(rule) {
		return b.WithFrom(builders.NewFromBuilder().WithForcedJWTAuthorization().Get())
	} else if processing.IsSecured(rule) {
		return b.WithFrom(builders.NewFromBuilder().WithOathkeeperProxySource().Get())
	}
	return b.WithFrom(builders.NewFromBuilder().WithIngressGatewaySource().Get())
}

// baseRuleBuilder returns RuleBuilder with To and From
func baseRuleBuilder(rule gatewayv1beta1.Rule) *builders.RuleBuilder {
	builder := builders.NewRuleBuilder()
	builder = withTo(builder, rule)
	builder = withFrom(builder, rule)

	return builder
}
