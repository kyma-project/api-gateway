package istio

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
)

const (
	audienceKey string = "request.auth.claims[aud]"
)

var (
	defaultScopeKeys = []string{"request.auth.claims[scp]", "request.auth.claims[scope]", "request.auth.claims[scopes]"}
)

// Newv1beta1AuthorizationPolicyProcessor returns a AuthorizationPolicyProcessor with the desired state handling specific for the Istio handler.
func Newv1beta1AuthorizationPolicyProcessor(config processing.ReconciliationConfig, log *logr.Logger, rule *gatewayv1beta1.APIRule) processors.AuthorizationPolicyProcessor {
	return processors.AuthorizationPolicyProcessor{
		ApiRule: rule,
		Creator: authorizationPolicyCreator{},
		Log:     log,
	}
}

type authorizationPolicyCreator struct{}

// Create returns the JwtAuthorization Policy using the configuration of the APIRule.
func (r authorizationPolicyCreator) Create(ctx context.Context, client client.Client, api *gatewayv1beta1.APIRule) (hashbasedstate.Desired, error) {
	state := hashbasedstate.NewDesired()
	if processing.RequiresAuthorizationPolicies(api) {
		for _, rule := range api.Spec.Rules {
			aps, err := generateAuthorizationPolicies(ctx, client, api, rule)
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
	}
	return state, nil
}

func generateAuthorizationPolicies(
	ctx context.Context,
	client client.Client,
	api *gatewayv1beta1.APIRule,
	rule gatewayv1beta1.Rule,
) (*securityv1beta1.AuthorizationPolicyList, error) {
	authorizationPolicyList := securityv1beta1.AuthorizationPolicyList{}
	ruleAuthorizations := rule.GetJwtIstioAuthorizations()

	if len(ruleAuthorizations) == 0 {
		// In case of no_auth, it will create an ALLOW AuthorizationPolicy bypassing any other AuthorizationPolicies.
		ap, err := generateAuthorizationPolicy(ctx, client, api, rule, &gatewayv1beta1.JwtAuthorization{})
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
		for indexInYaml, authorization := range ruleAuthorizations {
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

func generateAuthorizationPolicy(
	ctx context.Context,
	client client.Client,
	api *gatewayv1beta1.APIRule,
	rule gatewayv1beta1.Rule,
	authorization *gatewayv1beta1.JwtAuthorization,
) (*securityv1beta1.AuthorizationPolicy, error) {
	namePrefix := fmt.Sprintf("%s-", api.Name)
	namespace := helpers.FindServiceNamespace(api, &rule)

	spec, err := generateAuthorizationPolicySpec(ctx, client, api, rule, authorization)
	if err != nil {
		return nil, err
	}

	apBuilder := builders.NewAuthorizationPolicyBuilder().
		WithGenerateName(namePrefix).
		WithNamespace(namespace).
		WithSpec(builders.NewAuthorizationPolicySpecBuilder().FromAP(spec).Get()).
		WithLabel(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.Name, api.Namespace))

	return apBuilder.Get(), nil
}

func generateAuthorizationPolicySpec(
	ctx context.Context,
	client client.Client,
	api *gatewayv1beta1.APIRule,
	rule gatewayv1beta1.Rule,
	authorization *gatewayv1beta1.JwtAuthorization,
) (*v1beta1.AuthorizationPolicy, error) {
	var service *gatewayv1beta1.Service
	if rule.Service != nil {
		service = rule.Service
	} else {
		service = api.Spec.Service
	}

	labelSelector, err := helpers.GetLabelSelectorFromService(ctx, client, service, api, &rule)
	if err != nil {
		return nil, err
	}

	authorizationPolicySpecBuilder := builders.NewAuthorizationPolicySpecBuilder().WithSelector(labelSelector)

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

func withTo(b *builders.RuleBuilder, rule gatewayv1beta1.Rule) *builders.RuleBuilder {
	// APIRule and VirtualService supported a regex match. Since AuthorizationPolicy supports only prefix, suffix and wildcard
	// and we have clusters with "/.*" in APIRule, we need special handling of this case.
	path := rule.Path
	if rule.Path == "/.*" {
		path = "/*"
	}
	return b.WithTo(
		builders.NewToBuilder().
			WithOperation(builders.NewOperationBuilder().
				WithMethods(rule.Methods).WithPath(path).Get()).
			Get())
}

func withFrom(b *builders.RuleBuilder, rule gatewayv1beta1.Rule) *builders.RuleBuilder {
	if processing.IsJwtSecured(rule) {
		return b.WithFrom(builders.NewFromBuilder().WithForcedJWTAuthorization(rule.AccessStrategies).Get())
	} else if processing.IsSecuredByOathkeeper(rule) {
		return b.WithFrom(builders.NewFromBuilder().WithOathkeeperProxySource().Get())
	}
	return b.WithFrom(builders.NewFromBuilder().WithIngressGatewaySource().Get())
}

// baseRuleBuilder returns RuleBuilder with To and From.
func baseRuleBuilder(rule gatewayv1beta1.Rule) *builders.RuleBuilder {
	builder := builders.NewRuleBuilder()
	builder = withTo(builder, rule)
	builder = withFrom(builder, rule)

	return builder
}
