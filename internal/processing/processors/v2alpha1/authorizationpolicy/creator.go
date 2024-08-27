package authorizationpolicy

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	audienceKey string = "request.auth.claims[aud]"
)

var (
	defaultScopeKeys = []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"}
)

// Creator provides the creation of AuthorizationPolicy using the configuration in the given APIRule.
type Creator interface {
	Create(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule) (hashbasedstate.Desired, error)
}

type creator struct{}

// Create returns the AuthorizationPolicy using the configuration of the APIRule.
func (r creator) Create(ctx context.Context, client client.Client, apiRule *gatewayv2alpha1.APIRule) (hashbasedstate.Desired, error) {
	state := hashbasedstate.NewDesired()
	for _, rule := range apiRule.Spec.Rules {
		aps, err := generateAuthorizationPolicies(ctx, client, apiRule, rule)
		if err != nil {
			return state, err
		}

		for _, ap := range aps.Items {
			h := hashbasedstate.NewAuthorizationPolicy(ap)
			err := state.Add(&h)

			if err != nil {
				return state, err
			}
		}
	}
	return state, nil
}

func generateAuthorizationPolicies(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (*securityv1beta1.AuthorizationPolicyList, error) {
	authorizationPolicyList := securityv1beta1.AuthorizationPolicyList{}

	var jwtAuthorizations []*gatewayv2alpha1.JwtAuthorization
	if rule.Jwt != nil {
		jwtAuthorizations = rule.Jwt.Authorizations
	}

	if len(jwtAuthorizations) == 0 {
		// In case of NoAuth, it will create an ALLOW AuthorizationPolicy bypassing any other AuthorizationPolicies.
		ap, err := generateAuthorizationPolicy(ctx, client, api, rule, &gatewayv2alpha1.JwtAuthorization{})
		if err != nil {
			return &authorizationPolicyList, err
		}

		// If there is no other authorization we can safely assume that the index of this authorization in the array
		// in the yaml is 0.
		err = hashbasedstate.AddLabelsToAuthorizationPolicy(ap, 0)
		if err != nil {
			return &authorizationPolicyList, err
		}

		authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
	} else {
		for indexInYaml, authorization := range jwtAuthorizations {
			ap, err := generateAuthorizationPolicy(ctx, client, api, rule, authorization)
			if err != nil {
				return &authorizationPolicyList, err
			}

			err = hashbasedstate.AddLabelsToAuthorizationPolicy(ap, indexInYaml)
			if err != nil {
				return &authorizationPolicyList, err
			}

			authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
		}
	}

	return &authorizationPolicyList, nil
}

func generateAuthorizationPolicy(ctx context.Context, client client.Client, apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule, authorization *gatewayv2alpha1.JwtAuthorization) (*securityv1beta1.AuthorizationPolicy, error) {
	namePrefix := fmt.Sprintf("%s-", apiRule.ObjectMeta.Name)
	namespace, err := gatewayv2alpha1.FindServiceNamespace(apiRule, rule)
	if err != nil {
		return nil, fmt.Errorf("finding service namespace: %w", err)
	}

	spec, err := generateAuthorizationPolicySpec(ctx, client, apiRule, rule, authorization)
	if err != nil {
		return nil, err
	}

	apBuilder := builders.NewAuthorizationPolicyBuilder().
		WithGenerateName(namePrefix).
		WithNamespace(namespace).
		WithSpec(builders.NewAuthorizationPolicySpecBuilder().FromAP(spec).Get()).
		WithLabel(processing.OwnerLabel, fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace))

	return apBuilder.Get(), nil
}

func generateAuthorizationPolicySpec(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule, authorization *gatewayv2alpha1.JwtAuthorization) (*v1beta1.AuthorizationPolicy, error) {
	podSelector, err := gatewayv2alpha1.GetSelectorFromService(ctx, client, api, rule)
	if err != nil {
		return nil, err
	}

	authorizationPolicySpecBuilder := builders.NewAuthorizationPolicySpecBuilder().WithSelector(podSelector.Selector)

	// If RequiredScopes are configured, we need to generate a separate Rule for each scopeKey in defaultScopeKeys
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

	return authorizationPolicySpecBuilder.Get(), nil
}

func withTo(b *builders.RuleBuilder, rule gatewayv2alpha1.Rule) *builders.RuleBuilder {
	// APIRule and VirtualService supported a regex match. Since AuthorizationPolicy supports only prefix, suffix and wildcard,
	// and we have clusters with "/.*" in APIRule, we need special handling of this case.
	if rule.Path == "/.*" {
		return b.WithTo(
			builders.NewToBuilder().
				WithOperation(builders.NewOperationBuilder().
					WithMethodsV2alpha1(rule.Methods).WithPath("/*").Get()).
				Get())
	}
	return b.WithTo(
		builders.NewToBuilder().
			WithOperation(builders.NewOperationBuilder().
				WithMethodsV2alpha1(rule.Methods).WithPath(rule.Path).Get()).
			Get())
}

func withFrom(b *builders.RuleBuilder, rule gatewayv2alpha1.Rule) *builders.RuleBuilder {
	if rule.Jwt != nil {
		return b.WithFrom(builders.NewFromBuilder().WithForcedJWTAuthorizationV2alpha1(rule.Jwt.Authentications).Get())
	}

	return b.WithFrom(builders.NewFromBuilder().WithIngressGatewaySource().Get())
}

// baseRuleBuilder returns ruleBuilder with To and From
func baseRuleBuilder(rule gatewayv2alpha1.Rule) *builders.RuleBuilder {
	builder := builders.NewRuleBuilder()
	builder = withTo(builder, rule)
	builder = withFrom(builder, rule)

	return builder
}
