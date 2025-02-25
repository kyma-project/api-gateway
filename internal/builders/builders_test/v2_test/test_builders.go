package v2_test

import (
	gatewayv2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	"k8s.io/utils/ptr"
	"net/http"
)

type RuleBuilder struct {
	rule *gatewayv2.Rule
}

func (r *RuleBuilder) WithPath(path string) *RuleBuilder {
	r.rule.Path = path
	return r
}

func (r *RuleBuilder) WithTimeout(timeout uint32) *RuleBuilder {
	r.rule.Timeout = ptr.To(gatewayv2.Timeout(timeout))
	return r
}

func (r *RuleBuilder) WithService(name, namespace string, port uint32) *RuleBuilder {
	r.rule.Service = &gatewayv2.Service{
		Name:      &name,
		Namespace: &namespace,
		Port:      &port,
	}
	return r
}

func (r *RuleBuilder) WithMethods(methods ...gatewayv2.HttpMethod) *RuleBuilder {
	r.rule.Methods = methods
	return r
}

func (r *RuleBuilder) NoAuth() *RuleBuilder {
	r.rule.NoAuth = ptr.To(true)
	return r
}

func (r *RuleBuilder) WithJWTAuthn(issuer, jwksUri string, fromHeaders []*gatewayv2.JwtHeader, fromParams []string) *RuleBuilder {
	if r.rule.Jwt == nil {
		r.rule.Jwt = &gatewayv2.JwtConfig{}
	}
	r.rule.Jwt.Authentications = append(r.rule.Jwt.Authentications, &gatewayv2.JwtAuthentication{
		Issuer:      issuer,
		JwksUri:     jwksUri,
		FromHeaders: fromHeaders,
		FromParams:  fromParams,
	})

	return r
}

func (r *RuleBuilder) WithJWTAuthz(requiredScopes []string, audiences []string) *RuleBuilder {
	if r.rule.Jwt == nil {
		r.rule.Jwt = &gatewayv2.JwtConfig{}
	}

	r.rule.Jwt.Authorizations = append(r.rule.Jwt.Authorizations, &gatewayv2.JwtAuthorization{
		RequiredScopes: requiredScopes,
		Audiences:      audiences,
	})

	return r
}

func (r *RuleBuilder) WithRequest(rm *gatewayv2.Request) *RuleBuilder {
	r.rule.Request = rm
	return r
}

func (r *RuleBuilder) WithExtAuth(auth *gatewayv2.ExtAuth) *RuleBuilder {
	r.rule.ExtAuth = auth
	return r
}

func NewRuleBuilder() *RuleBuilder {
	return &RuleBuilder{
		rule: &gatewayv2.Rule{},
	}
}

func (r *RuleBuilder) Build() *gatewayv2.Rule {
	return r.rule
}

type RequestBuilder struct {
	request *gatewayv2.Request
}

func (m *RequestBuilder) WithHeaders(headers map[string]string) *RequestBuilder {
	m.request.Headers = headers

	return m
}

func (m *RequestBuilder) WithCookies(cookies map[string]string) *RequestBuilder {
	m.request.Cookies = cookies

	return m
}

func NewRequestModifier() *RequestBuilder {
	return &RequestBuilder{
		request: &gatewayv2.Request{},
	}
}

func (m *RequestBuilder) Build() *gatewayv2.Request {
	return m.request
}

type ApiRuleBuilder struct {
	apiRule *gatewayv2.APIRule
}

func (a *ApiRuleBuilder) WithHost(host string) *ApiRuleBuilder {
	a.apiRule.Spec.Hosts = append(a.apiRule.Spec.Hosts, ptr.To(gatewayv2.Host(host)))
	return a
}

func (a *ApiRuleBuilder) WithHosts(hosts ...string) *ApiRuleBuilder {
	for _, host := range hosts {
		a.WithHost(host)
	}
	return a
}

func (a *ApiRuleBuilder) WithService(name, namespace string, port uint32) *ApiRuleBuilder {
	a.apiRule.Spec.Service = &gatewayv2.Service{
		Name:      &name,
		Namespace: &namespace,
		Port:      &port,
	}
	return a
}

func (a *ApiRuleBuilder) WithGateway(gateway string) *ApiRuleBuilder {
	a.apiRule.Spec.Gateway = ptr.To(gateway)
	return a
}

func (a *ApiRuleBuilder) WithCORSPolicy(policy gatewayv2.CorsPolicy) *ApiRuleBuilder {
	a.apiRule.Spec.CorsPolicy = &policy
	return a
}

func (a *ApiRuleBuilder) WithTimeout(timeout uint32) *ApiRuleBuilder {
	a.apiRule.Spec.Timeout = ptr.To(gatewayv2.Timeout(timeout))
	return a
}

func (a *ApiRuleBuilder) WithRule(rule gatewayv2.Rule) *ApiRuleBuilder {
	a.apiRule.Spec.Rules = append(a.apiRule.Spec.Rules, rule)
	return a
}

func (a *ApiRuleBuilder) WithRules(rules ...*gatewayv2.Rule) *ApiRuleBuilder {
	for _, rule := range rules {
		a.WithRule(*rule)
	}
	return a
}

func (a *ApiRuleBuilder) Build() *gatewayv2.APIRule {
	return a.apiRule
}

func NewAPIRuleBuilder() *ApiRuleBuilder {
	return &ApiRuleBuilder{
		apiRule: &gatewayv2.APIRule{},
	}
}

// NewAPIRuleBuilderWithDummyDataWithNoAuthRule returns an APIRuleBuilder pre-filled with placeholder data:
//
// Host: example-host.example.com
//
// Gateway: example-namespace/example-gateway
//
// Service: example-namespace/example-service:8080
//
// Rule: GET /
//
// Strategy: NoAuth
func NewAPIRuleBuilderWithDummyDataWithNoAuthRule() *ApiRuleBuilder {
	return NewAPIRuleBuilder().
		WithHost("example-host.example.com").
		WithGateway("example-namespace/example-gateway").
		WithService("example-service", "example-namespace", 8080).
		WithRule(*NewRuleBuilder().WithMethods(http.MethodGet).WithPath("/").NoAuth().Build())
}

func NewAPIRuleBuilderWithDummyData() *ApiRuleBuilder {
	return NewAPIRuleBuilder().
		WithHost("example-host.example.com").
		WithGateway("example-namespace/example-gateway").
		WithService("example-service", "example-namespace", 8080)
}

type CorsPolicyBuilder struct {
	policy gatewayv2.CorsPolicy
}

func (c *CorsPolicyBuilder) WithAllowOrigins(origins []map[string]string) *CorsPolicyBuilder {
	c.policy.AllowOrigins = origins
	return c
}

func (c *CorsPolicyBuilder) WithAllowMethods(methods []string) *CorsPolicyBuilder {
	c.policy.AllowMethods = methods
	return c
}

func (c *CorsPolicyBuilder) WithAllowHeaders(headers []string) *CorsPolicyBuilder {
	c.policy.AllowHeaders = headers
	return c
}

func (c *CorsPolicyBuilder) WithExposeHeaders(headers []string) *CorsPolicyBuilder {
	c.policy.ExposeHeaders = headers
	return c
}

func (c *CorsPolicyBuilder) WithMaxAge(maxAge uint64) *CorsPolicyBuilder {
	c.policy.MaxAge = &maxAge
	return c
}

func (c *CorsPolicyBuilder) WithAllowCredentials(allow bool) *CorsPolicyBuilder {
	c.policy.AllowCredentials = &allow
	return c
}

func NewCorsPolicyBuilder() *CorsPolicyBuilder {
	return &CorsPolicyBuilder{
		policy: gatewayv2.CorsPolicy{},
	}
}

func (c *CorsPolicyBuilder) Build() gatewayv2.CorsPolicy {
	return c.policy
}

type ExtAuthBuilder struct {
	extAuth *gatewayv2.ExtAuth
}

func NewExtAuthBuilder() *ExtAuthBuilder {
	return &ExtAuthBuilder{
		extAuth: &gatewayv2.ExtAuth{},
	}
}

func (e *ExtAuthBuilder) Build() *gatewayv2.ExtAuth {
	return e.extAuth
}

func (e *ExtAuthBuilder) WithAuthorizers(auths ...string) *ExtAuthBuilder {
	e.extAuth.ExternalAuthorizers = append(e.extAuth.ExternalAuthorizers, auths...)
	return e
}

func (e *ExtAuthBuilder) WithRestriction(config *gatewayv2.JwtConfig) *ExtAuthBuilder {
	e.extAuth.Restrictions = config
	return e
}

func (e *ExtAuthBuilder) WithRestrictionAuthorization(config *gatewayv2.JwtAuthorization) *ExtAuthBuilder {
	if e.extAuth.Restrictions == nil {
		e.extAuth.Restrictions = &gatewayv2.JwtConfig{}
	}

	e.extAuth.Restrictions.Authorizations = append(e.extAuth.Restrictions.Authorizations, config)
	return e
}

func (e *ExtAuthBuilder) WithRestrictionAuthentication(config *gatewayv2.JwtAuthentication) *ExtAuthBuilder {
	if e.extAuth.Restrictions == nil {
		e.extAuth.Restrictions = &gatewayv2.JwtConfig{}
	}
	e.extAuth.Restrictions.Authentications = append(e.extAuth.Restrictions.Authentications, config)
	return e
}
