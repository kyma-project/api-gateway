package virtualservice_test

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"k8s.io/utils/ptr"
	"net/http"
)

type ruleBuilder struct {
	rule *gatewayv2alpha1.Rule
}

func (r *ruleBuilder) WithPath(path string) *ruleBuilder {
	r.rule.Path = path
	return r
}

func (r *ruleBuilder) WithTimeout(timeout uint32) *ruleBuilder {
	r.rule.Timeout = ptr.To(gatewayv2alpha1.Timeout(timeout))
	return r
}

func (r *ruleBuilder) WithService(name, namespace string, port uint32) *ruleBuilder {
	r.rule.Service = &gatewayv2alpha1.Service{
		Name:      &name,
		Namespace: &namespace,
		Port:      &port,
	}
	return r
}

func (r *ruleBuilder) WithMethods(methods ...gatewayv2alpha1.HttpMethod) *ruleBuilder {
	r.rule.Methods = methods
	return r
}

func (r *ruleBuilder) NoAuth() *ruleBuilder {
	r.rule.NoAuth = ptr.To(true)
	return r
}

func (r *ruleBuilder) WithJWTAuthn(issuer, jwksUri string, fromHeaders []*gatewayv2alpha1.JwtHeader, fromParams []string) *ruleBuilder {
	if r.rule.Jwt == nil {
		r.rule.Jwt = &gatewayv2alpha1.JwtConfig{}
	}
	r.rule.Jwt.Authentications = append(r.rule.Jwt.Authentications, &gatewayv2alpha1.JwtAuthentication{
		Issuer:      issuer,
		JwksUri:     jwksUri,
		FromHeaders: fromHeaders,
		FromParams:  fromParams,
	})

	return r
}

func (r *ruleBuilder) WithJWTAuthz(requiredScopes []string, audiences []string) *ruleBuilder {
	if r.rule.Jwt == nil {
		r.rule.Jwt = &gatewayv2alpha1.JwtConfig{}
	}

	r.rule.Jwt.Authorizations = append(r.rule.Jwt.Authorizations, &gatewayv2alpha1.JwtAuthorization{
		RequiredScopes: requiredScopes,
		Audiences:      audiences,
	})

	return r
}

func (r *ruleBuilder) WithRequest(rm *gatewayv2alpha1.Request) *ruleBuilder {
	r.rule.Request = rm
	return r
}

func newRuleBuilder() *ruleBuilder {
	return &ruleBuilder{
		rule: &gatewayv2alpha1.Rule{},
	}
}

func (r *ruleBuilder) Build() *gatewayv2alpha1.Rule {
	return r.rule
}

type requestBuilder struct {
	request *gatewayv2alpha1.Request
}

func (m *requestBuilder) WithHeaders(headers map[string]string) *requestBuilder {
	m.request.Headers = headers

	return m
}

func (m *requestBuilder) WithCookies(cookies map[string]string) *requestBuilder {
	m.request.Cookies = cookies

	return m
}

func newRequestModifier() *requestBuilder {
	return &requestBuilder{
		request: &gatewayv2alpha1.Request{},
	}
}

func (m *requestBuilder) Build() *gatewayv2alpha1.Request {
	return m.request
}

type apiRuleBuilder struct {
	apiRule *gatewayv2alpha1.APIRule
}

func (a *apiRuleBuilder) WithHost(host string) *apiRuleBuilder {
	a.apiRule.Spec.Hosts = append(a.apiRule.Spec.Hosts, ptr.To(gatewayv2alpha1.Host(host)))
	return a
}

func (a *apiRuleBuilder) WithHosts(hosts ...string) *apiRuleBuilder {
	for _, host := range hosts {
		a.WithHost(host)
	}
	return a
}

func (a *apiRuleBuilder) WithService(name, namespace string, port uint32) *apiRuleBuilder {
	a.apiRule.Spec.Service = &gatewayv2alpha1.Service{
		Name:      &name,
		Namespace: &namespace,
		Port:      &port,
	}
	return a
}

func (a *apiRuleBuilder) WithGateway(gateway string) *apiRuleBuilder {
	a.apiRule.Spec.Gateway = ptr.To(gateway)
	return a
}

func (a *apiRuleBuilder) WithCORSPolicy(policy gatewayv2alpha1.CorsPolicy) *apiRuleBuilder {
	a.apiRule.Spec.CorsPolicy = &policy
	return a
}

func (a *apiRuleBuilder) WithTimeout(timeout uint32) *apiRuleBuilder {
	a.apiRule.Spec.Timeout = ptr.To(gatewayv2alpha1.Timeout(timeout))
	return a
}

func (a *apiRuleBuilder) WithRule(rule gatewayv2alpha1.Rule) *apiRuleBuilder {
	a.apiRule.Spec.Rules = append(a.apiRule.Spec.Rules, rule)
	return a
}

func (a *apiRuleBuilder) WithRules(rules ...*gatewayv2alpha1.Rule) *apiRuleBuilder {
	for _, rule := range rules {
		a.WithRule(*rule)
	}
	return a
}

func (a *apiRuleBuilder) Build() *gatewayv2alpha1.APIRule {
	return a.apiRule
}

func newAPIRuleBuilder() *apiRuleBuilder {
	return &apiRuleBuilder{
		apiRule: &gatewayv2alpha1.APIRule{},
	}
}

// newAPIRuleBuilderWithDummyDataWithNoAuthRule returns an APIRuleBuilder pre-filled with placeholder data:
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
func newAPIRuleBuilderWithDummyDataWithNoAuthRule() *apiRuleBuilder {
	return newAPIRuleBuilder().
		WithHost("example-host.example.com").
		WithGateway("example-namespace/example-gateway").
		WithService("example-service", "example-namespace", 8080).
		WithRule(*newRuleBuilder().WithMethods(http.MethodGet).WithPath("/").NoAuth().Build())
}

func newAPIRuleBuilderWithDummyData() *apiRuleBuilder {
	return newAPIRuleBuilder().
		WithHost("example-host.example.com").
		WithGateway("example-namespace/example-gateway").
		WithService("example-service", "example-namespace", 8080)
}

type corsPolicyBuilder struct {
	policy gatewayv2alpha1.CorsPolicy
}

func (c *corsPolicyBuilder) WithAllowOrigins(origins []map[string]string) *corsPolicyBuilder {
	c.policy.AllowOrigins = origins
	return c
}

func (c *corsPolicyBuilder) WithAllowMethods(methods []string) *corsPolicyBuilder {
	c.policy.AllowMethods = methods
	return c
}

func (c *corsPolicyBuilder) WithAllowHeaders(headers []string) *corsPolicyBuilder {
	c.policy.AllowHeaders = headers
	return c
}

func (c *corsPolicyBuilder) WithExposeHeaders(headers []string) *corsPolicyBuilder {
	c.policy.ExposeHeaders = headers
	return c
}

func (c *corsPolicyBuilder) WithMaxAge(maxAge uint64) *corsPolicyBuilder {
	c.policy.MaxAge = &maxAge
	return c
}

func (c *corsPolicyBuilder) WithAllowCredentials(allow bool) *corsPolicyBuilder {
	c.policy.AllowCredentials = &allow
	return c
}

func newCorsPolicyBuilder() *corsPolicyBuilder {
	return &corsPolicyBuilder{
		policy: gatewayv2alpha1.CorsPolicy{},
	}
}

func (c *corsPolicyBuilder) Build() gatewayv2alpha1.CorsPolicy {
	return c.policy
}
