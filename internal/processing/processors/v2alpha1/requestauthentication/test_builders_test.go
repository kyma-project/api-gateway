package requestauthentication_test

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"net/http"
)

type ruleBuilder struct {
	rule *gatewayv2alpha1.Rule
}

func newRuleBuilder() *ruleBuilder {
	return &ruleBuilder{
		rule: &gatewayv2alpha1.Rule{},
	}
}

func (b *ruleBuilder) withPath(path string) *ruleBuilder {
	b.rule.Path = path
	return b
}

func (b *ruleBuilder) addMethods(methods ...gatewayv2alpha1.HttpMethod) *ruleBuilder {
	b.rule.Methods = append(b.rule.Methods, methods...)
	return b
}

func (b *ruleBuilder) withServiceName(name string) *ruleBuilder {
	if b.rule.Service == nil {
		b.rule.Service = &gatewayv2alpha1.Service{}
	}

	b.rule.Service.Name = ptr.To(name)
	return b
}

func (b *ruleBuilder) withServiceNamespace(namespace string) *ruleBuilder {
	if b.rule.Service == nil {
		b.rule.Service = &gatewayv2alpha1.Service{}
	}

	b.rule.Service.Namespace = ptr.To(namespace)
	return b
}

func (b *ruleBuilder) withServicePort(port uint32) *ruleBuilder {
	if b.rule.Service == nil {
		b.rule.Service = &gatewayv2alpha1.Service{}
	}

	b.rule.Service.Port = ptr.To(port)
	return b
}

func (b *ruleBuilder) withNoAuth() *ruleBuilder {
	b.rule.NoAuth = ptr.To(true)
	return b
}

func (b *ruleBuilder) addJwtAuthentication(issuer, jwksUri string) *ruleBuilder {
	auth := &gatewayv2alpha1.JwtAuthentication{
		Issuer:  issuer,
		JwksUri: jwksUri,
	}

	if b.rule.Jwt == nil {
		b.rule.Jwt = &gatewayv2alpha1.JwtConfig{}
	}

	b.rule.Jwt.Authentications = append(b.rule.Jwt.Authentications, auth)
	return b
}

func (b *ruleBuilder) build() *gatewayv2alpha1.Rule {
	return b.rule
}

/*
newJwtRuleBuilderWithDummyData returns a ruleBuilder pre-filled with placeholder data:

Path: /

Methods: GET

Service: example-namespace/example-service:8080

JWT Authentication: https://oauth2.example.com/, https://oauth2.example.com/.well-known/jwks.json
*/
func newJwtRuleBuilderWithDummyData() *ruleBuilder {
	return newRuleBuilder().
		withPath("/").
		addMethods(http.MethodGet).
		withServiceName("example-service").
		withServiceNamespace("example-namespace").
		withServicePort(8080).
		addJwtAuthentication("https://oauth2.example.com/", "https://oauth2.example.com/.well-known/jwks.json")
}

/*
newNoAuthRuleBuilderWithDummyData returns a ruleBuilder pre-filled with placeholder data:

Path: /

Methods: GET

Service: example-namespace/example-service:8080

NoAuth: true
*/
func newNoAuthRuleBuilderWithDummyData() *ruleBuilder {
	return newRuleBuilder().
		withPath("/").
		addMethods(http.MethodGet).
		withServiceName("example-service").
		withServiceNamespace("example-namespace").
		withServicePort(8080).
		withNoAuth()
}

type apiRuleBuilder struct {
	apiRule *gatewayv2alpha1.APIRule
}

func (a *apiRuleBuilder) withHost(host string) *apiRuleBuilder {
	a.apiRule.Spec.Hosts = append(a.apiRule.Spec.Hosts, ptr.To(gatewayv2alpha1.Host(host)))
	return a
}

func (a *apiRuleBuilder) withServiceName(name string) *apiRuleBuilder {
	if a.apiRule.Spec.Service == nil {
		a.apiRule.Spec.Service = &gatewayv2alpha1.Service{}
	}

	a.apiRule.Spec.Service.Name = ptr.To(name)
	return a
}

func (a *apiRuleBuilder) withServiceNamespace(namespace string) *apiRuleBuilder {
	if a.apiRule.Spec.Service == nil {
		a.apiRule.Spec.Service = &gatewayv2alpha1.Service{}
	}

	a.apiRule.Spec.Service.Namespace = ptr.To(namespace)
	return a
}

func (a *apiRuleBuilder) withServicePort(port uint32) *apiRuleBuilder {
	if a.apiRule.Spec.Service == nil {
		a.apiRule.Spec.Service = &gatewayv2alpha1.Service{}
	}

	a.apiRule.Spec.Service.Port = ptr.To(port)
	return a
}

func (a *apiRuleBuilder) withGateway(gateway string) *apiRuleBuilder {
	a.apiRule.Spec.Gateway = ptr.To(gateway)
	return a
}

func (a *apiRuleBuilder) withName(name string) *apiRuleBuilder {
	a.apiRule.ObjectMeta.Name = name
	return a
}

func (a *apiRuleBuilder) withNamespace(namespace string) *apiRuleBuilder {
	a.apiRule.ObjectMeta.Namespace = namespace
	return a
}

func (a *apiRuleBuilder) withRule(rule gatewayv2alpha1.Rule) *apiRuleBuilder {
	a.apiRule.Spec.Rules = append(a.apiRule.Spec.Rules, rule)
	return a
}

func (a *apiRuleBuilder) withRules(rules ...*gatewayv2alpha1.Rule) *apiRuleBuilder {
	for _, rule := range rules {
		a.withRule(*rule)
	}
	return a
}

func (a *apiRuleBuilder) build() *gatewayv2alpha1.APIRule {
	return a.apiRule
}

func newAPIRuleBuilder() *apiRuleBuilder {
	return &apiRuleBuilder{
		apiRule: &gatewayv2alpha1.APIRule{},
	}
}

/*
newAPIRuleBuilderWithDummyDataWithNoAuthRule returns an APIRuleBuilder pre-filled with placeholder data:

Name: test-apirule

Namespace: example-namespace

Host: example-host.example.com

Gateway: example-namespace/example-gateway

Service: example-namespace/example-service:8080
*/
func newAPIRuleBuilderWithDummyData() *apiRuleBuilder {
	return newAPIRuleBuilder().
		withName(apiRuleName).
		withNamespace(apiRuleNamespace).
		withHost("example-host.example.com").
		withGateway("example-namespace/example-gateway").
		withServiceName("example-service").
		withServiceNamespace("example-namespace").
		withServicePort(8080)
}

type serviceBuilder struct {
	service *corev1.Service
}

func newServiceBuilder() *serviceBuilder {
	return &serviceBuilder{
		service: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{},
		},
	}
}

func (b *serviceBuilder) withName(name string) *serviceBuilder {
	b.service.ObjectMeta.Name = name
	return b
}

func (b *serviceBuilder) withNamespace(namespace string) *serviceBuilder {
	b.service.ObjectMeta.Namespace = namespace
	return b
}

func (b *serviceBuilder) addSelector(key, value string) *serviceBuilder {
	if b.service.Spec.Selector == nil {
		b.service.Spec.Selector = map[string]string{}
	}

	b.service.Spec.Selector[key] = value
	return b
}

func (b *serviceBuilder) build() *corev1.Service {
	return b.service
}

/*
newServiceBuilderWithDummyData returns a serviceBuilder pre-filled with placeholder data:

Name: example-service

Namespace: example-namespace

Selector: app=example-service
*/
func newServiceBuilderWithDummyData() *serviceBuilder {
	return newServiceBuilder().
		withName("example-service").
		withNamespace("example-namespace").
		addSelector("app", "example-service")

}
