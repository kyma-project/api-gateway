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

type creator struct {
	// Controls that requests to Ory Oathkeeper are also permitted when
	// migrating from APIRule v1beta1 to v2alpha1.
	oryPassthrough          bool
	disallowInternalTraffic bool
}

// Create returns the AuthorizationPolicy using the configuration of the APIRule.
func (r creator) Create(ctx context.Context, client client.Client, apiRule *gatewayv2alpha1.APIRule) (hashbasedstate.Desired, error) {
	state := hashbasedstate.NewDesired()
	selectorAllowed := make(map[gatewayv2alpha1.PodSelector]bool)
	for _, rule := range apiRule.Spec.Rules {
		selector, err := gatewayv2alpha1.GetSelectorFromService(ctx, client, apiRule, rule)
		if err != nil {
			return state, err
		}
		var aps *securityv1beta1.AuthorizationPolicyList
		if _, ok := selectorAllowed[selector]; ok || r.disallowInternalTraffic {
			aps, err = r.generateAuthorizationPolicies(ctx, client, apiRule, rule, false)
			if err != nil {
				return state, err
			}
		} else {
			aps, err = r.generateAuthorizationPolicies(ctx, client, apiRule, rule, true)
			if err != nil {
				return state, err
			}
			selectorAllowed[selector] = true
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

func (r creator) generateAllowForInternalTraffic(ctx context.Context, k8sClient client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (*securityv1beta1.AuthorizationPolicy, error) {
	podSelector, err := gatewayv2alpha1.GetSelectorFromService(ctx, k8sClient, api, rule)
	if err != nil {
		return nil, err
	}

	apBuilder, err := baseAuthorizationPolicyBuilder(api, rule)
	if err != nil {
		return nil, fmt.Errorf("error creating base AuthorizationPolicy builder: %w", err)
	}

	apBuilder.WithSpec(
		builders.NewAuthorizationPolicySpecBuilder().
			WithSelector(podSelector.Selector).
			WithAction(v1beta1.AuthorizationPolicy_ALLOW).
			WithRule(builders.NewRuleBuilder().
				WithFrom(
					builders.NewFromBuilder().
						ExcludingIngressGatewaySource().
						Get()).
				Get()).
			Get())

	return apBuilder.Get(), nil
}

func (r creator) generateAuthorizationPolicies(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule, allowInternalTraffic bool) (*securityv1beta1.AuthorizationPolicyList, error) {
	authorizationPolicyList := securityv1beta1.AuthorizationPolicyList{}

	var jwtAuthorizations []*gatewayv2alpha1.JwtAuthorization

	baseHashIndex := 0
	switch {
	case rule.Jwt != nil:
		jwtAuthorizations = append(jwtAuthorizations, rule.Jwt.Authorizations...)
	case rule.ExtAuth != nil:
		baseHashIndex = len(rule.ExtAuth.ExternalAuthorizers)
		if rule.ExtAuth.Restrictions != nil {
			jwtAuthorizations = append(jwtAuthorizations, rule.ExtAuth.Restrictions.Authorizations...)
		}
		policies, err := r.generateExtAuthAuthorizationPolicies(ctx, client, api, rule)
		if err != nil {
			return &authorizationPolicyList, err
		}

		authorizationPolicyList.Items = append(authorizationPolicyList.Items, policies...)
	}

	if len(jwtAuthorizations) == 0 {
		ap, err := r.generateAuthorizationPolicyForEmptyAuthorizations(ctx, client, api, rule)
		if err != nil {
			return &authorizationPolicyList, err
		}

		err = hashbasedstate.AddLabelsToAuthorizationPolicy(ap, baseHashIndex)
		if err != nil {
			return &authorizationPolicyList, err
		}

		authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
	} else {
		for indexInYaml, authorization := range jwtAuthorizations {
			ap, err := r.generateAuthorizationPolicy(ctx, client, api, rule, authorization)
			if err != nil {
				return &authorizationPolicyList, err
			}

			err = hashbasedstate.AddLabelsToAuthorizationPolicy(ap, indexInYaml+baseHashIndex)
			if err != nil {
				return &authorizationPolicyList, err
			}

			authorizationPolicyList.Items = append(authorizationPolicyList.Items, ap)
		}
	}

	if allowInternalTraffic {
		internalTrafficAp, err := r.generateAllowForInternalTraffic(ctx, client, api, rule)
		if err != nil {
			return &authorizationPolicyList, err
		}

		if internalTrafficAp != nil {
			err = hashbasedstate.AddLabelsToAuthorizationPolicy(internalTrafficAp, baseHashIndex+1+len(jwtAuthorizations))
			if err != nil {
				return &authorizationPolicyList, err
			}
			authorizationPolicyList.Items = append(authorizationPolicyList.Items, internalTrafficAp)
		}
	}

	return &authorizationPolicyList, nil
}

func (r creator) generateExtAuthAuthorizationPolicies(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (authorizationPolicyList []*securityv1beta1.AuthorizationPolicy, _ error) {
	for i, authorizer := range rule.ExtAuth.ExternalAuthorizers {
		policy, err := r.generateExtAuthAuthorizationPolicy(ctx, client, api, rule, authorizer)
		if err != nil {
			return authorizationPolicyList, err
		}

		err = hashbasedstate.AddLabelsToAuthorizationPolicy(policy, i)
		if err != nil {
			return authorizationPolicyList, err
		}

		authorizationPolicyList = append(authorizationPolicyList, policy)
	}

	return authorizationPolicyList, nil
}

func (r creator) generateAuthorizationPolicyForEmptyAuthorizations(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (*securityv1beta1.AuthorizationPolicy, error) {
	// In case of NoAuth, it will create an ALLOW AuthorizationPolicy bypassing any other AuthorizationPolicies.
	ap, err := r.generateAuthorizationPolicy(ctx, client, api, rule, &gatewayv2alpha1.JwtAuthorization{})
	if err != nil {
		return nil, err
	}

	// If there is no other authorization, we can safely assume that the index of this authorization in the array
	// in the YAML is 0.
	if err := hashbasedstate.AddLabelsToAuthorizationPolicy(ap, 0); err != nil {
		return nil, err
	}

	return ap, nil
}

func baseAuthorizationPolicyBuilder(apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (*builders.AuthorizationPolicyBuilder, error) {
	namePrefix := fmt.Sprintf("%s-", apiRule.ObjectMeta.Name)
	namespace, err := gatewayv2alpha1.FindServiceNamespace(apiRule, rule)
	if err != nil {
		return nil, fmt.Errorf("finding service namespace: %w", err)
	}

	return builders.NewAuthorizationPolicyBuilder().
			WithGenerateName(namePrefix).
			WithNamespace(namespace).
			WithLabel(processing.OwnerLabel, fmt.Sprintf("%s.%s", apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace)),
		nil
}

func (r creator) generateExtAuthAuthorizationPolicy(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule, authorizerName string) (*securityv1beta1.AuthorizationPolicy, error) {
	spec, err := r.generateExtAuthAuthorizationPolicySpec(ctx, client, api, rule, authorizerName)
	if err != nil {
		return nil, err
	}

	apBuilder, err := baseAuthorizationPolicyBuilder(api, rule)
	if err != nil {
		return nil, fmt.Errorf("error creating base AuthorizationPolicy builder: %w", err)
	}

	apBuilder.
		WithSpec(builders.NewAuthorizationPolicySpecBuilder().FromAP(spec).Get())

	return apBuilder.Get(), nil
}

func (r creator) generateAuthorizationPolicy(ctx context.Context, client client.Client, apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule, authorization *gatewayv2alpha1.JwtAuthorization) (*securityv1beta1.AuthorizationPolicy, error) {
	spec, err := r.generateAuthorizationPolicySpec(ctx, client, apiRule, rule, authorization)
	if err != nil {
		return nil, err
	}

	apBuilder, err := baseAuthorizationPolicyBuilder(apiRule, rule)
	if err != nil {
		return nil, fmt.Errorf("error creating base AuthorizationPolicy builder: %w", err)
	}

	apBuilder.WithSpec(
		builders.NewAuthorizationPolicySpecBuilder().
			FromAP(spec).
			Get())

	return apBuilder.Get(), nil
}

func (r creator) generateExtAuthAuthorizationPolicySpec(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule, providerName string) (*v1beta1.AuthorizationPolicy, error) {
	podSelector, err := gatewayv2alpha1.GetSelectorFromService(ctx, client, api, rule)
	if err != nil {
		return nil, err
	}

	authorizationPolicySpecBuilder := builders.NewAuthorizationPolicySpecBuilder().WithSelector(podSelector.Selector)
	return authorizationPolicySpecBuilder.
		WithAction(v1beta1.AuthorizationPolicy_CUSTOM).
		WithProvider(providerName).
		WithRule(baseExtAuthRuleBuilder(rule).Get()).
		Get(), nil
}

func (r creator) generateAuthorizationPolicySpec(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule, authorization *gatewayv2alpha1.JwtAuthorization) (*v1beta1.AuthorizationPolicy, error) {
	podSelector, err := gatewayv2alpha1.GetSelectorFromService(ctx, client, api, rule)
	if err != nil {
		return nil, err
	}

	authorizationPolicySpecBuilder := builders.NewAuthorizationPolicySpecBuilder().
		WithSelector(podSelector.Selector)

	// If RequiredScopes are configured, we need to generate a separate Rule for each scopeKey in defaultScopeKeys
	if len(authorization.RequiredScopes) > 0 {
		for _, scopeKey := range defaultScopeKeys {
			ruleBuilder := baseRuleBuilder(rule, r.oryPassthrough)
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
		ruleBuilder := baseRuleBuilder(rule, r.oryPassthrough)
		for _, aud := range authorization.Audiences {
			ruleBuilder.WithWhenCondition(
				builders.NewConditionBuilder().WithKey(audienceKey).WithValues([]string{aud}).Get())
		}
		authorizationPolicySpecBuilder.WithRule(ruleBuilder.Get())
	}

	return authorizationPolicySpecBuilder.Get(), nil
}

// standardizeRulePath converts wildcard `/*` path to post Istio 1.22 Envoy template format `/{**}`.
func standardizeRulePath(path string) string {
	if path == "/*" {
		return "/{**}"
	}
	return path
}

func withTo(b *builders.RuleBuilder, rule gatewayv2alpha1.Rule) *builders.RuleBuilder {
	return b.WithTo(
		builders.NewToBuilder().
			WithOperation(builders.NewOperationBuilder().
				WithMethodsV2alpha1(rule.Methods).WithPath(standardizeRulePath(rule.Path)).Get()).
			Get())
}

func withFrom(b *builders.RuleBuilder, rule gatewayv2alpha1.Rule, oryPassthrough bool) *builders.RuleBuilder {
	if rule.Jwt != nil {
		return b.WithFrom(builders.NewFromBuilder().WithForcedJWTAuthorizationV2alpha1(rule.Jwt.Authentications).Get())
	} else if rule.ExtAuth != nil && rule.ExtAuth.Restrictions != nil {
		return b.WithFrom(builders.NewFromBuilder().WithForcedJWTAuthorizationV2alpha1(rule.ExtAuth.Restrictions.Authentications).Get())
	}

	if oryPassthrough {
		b.WithFrom(builders.NewFromBuilder().WithOathkeeperProxySource().Get())
	}

	return b.WithFrom(builders.NewFromBuilder().WithIngressGatewaySource().Get())
}

// baseExtAuthRuleBuilder returns ruleBuilder with To
func baseExtAuthRuleBuilder(rule gatewayv2alpha1.Rule) *builders.RuleBuilder {
	builder := builders.NewRuleBuilder()
	builder = withTo(builder, rule)

	return builder
}

// baseRuleBuilder returns ruleBuilder with To and From
func baseRuleBuilder(rule gatewayv2alpha1.Rule, oryPassthrough bool) *builders.RuleBuilder {
	builder := builders.NewRuleBuilder()
	builder = withTo(builder, rule)
	builder = withFrom(builder, rule, oryPassthrough)

	return builder
}
