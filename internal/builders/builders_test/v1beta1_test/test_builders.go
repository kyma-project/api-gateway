package v1beta1_test

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
)

type APIRuleBuilder struct {
	rule *gatewayv1beta1.APIRule
}

func NewAPIRuleBuilder() *APIRuleBuilder {
	return &APIRuleBuilder{
		rule: &gatewayv1beta1.APIRule{},
	}
}

func NewAPIRuleBuilderWithDummyData() *APIRuleBuilder {
	return NewAPIRuleBuilder().
		WithGateway("example/example").
		WithHost("example.com").
		WithService("example-service", "example-namespace", 8080)
}

func (r *APIRuleBuilder) Build() *gatewayv1beta1.APIRule {
	return r.rule
}

func (r *APIRuleBuilder) WithGateway(gateway string) *APIRuleBuilder {
	r.rule.Spec.Gateway = &gateway
	return r
}

func (r *APIRuleBuilder) WithService(name, namespace string, port uint32) *APIRuleBuilder {
	r.rule.Spec.Service = &gatewayv1beta1.Service{
		Name:      &name,
		Namespace: &namespace,
		Port:      &port,
	}
	return r
}

func (r *APIRuleBuilder) WithHost(host string) *APIRuleBuilder {
	r.rule.Spec.Host = &host
	return r
}

func (r *APIRuleBuilder) WithTimeout(timeout gatewayv1beta1.Timeout) *APIRuleBuilder {
	r.rule.Spec.Timeout = &timeout
	return r
}

func (r *APIRuleBuilder) WithCorsPolicy(policy gatewayv1beta1.CorsPolicy) *APIRuleBuilder {
	r.rule.Spec.CorsPolicy = &policy
	return r
}

func (r *APIRuleBuilder) WithRule(rule gatewayv1beta1.Rule) *APIRuleBuilder {
	r.rule.Spec.Rules = append(r.rule.Spec.Rules, rule)
	return r
}

type RuleBuilder struct {
	rule *gatewayv1beta1.Rule
}

func NewRuleBuilder() *RuleBuilder {
	return &RuleBuilder{
		rule: &gatewayv1beta1.Rule{},
	}
}

func (r *RuleBuilder) WithPath(path string) *RuleBuilder {
	r.rule.Path = path
	return r
}

func (r *RuleBuilder) WithMethods(methods ...gatewayv1beta1.HttpMethod) *RuleBuilder {
	r.rule.Methods = methods
	return r
}

func (r *RuleBuilder) WithHandler(handler gatewayv1beta1.Handler) *RuleBuilder {
	r.rule.AccessStrategies = append(r.rule.AccessStrategies, &gatewayv1beta1.Authenticator{Handler: &handler})
	return r
}

func (r *RuleBuilder) WithService(name, namespace string, port uint32) *RuleBuilder {
	r.rule.Service = &gatewayv1beta1.Service{
		Name:      &name,
		Namespace: &namespace,
		Port:      &port,
	}
	return r
}

func (r *RuleBuilder) Build() *gatewayv1beta1.Rule {
	return r.rule
}
