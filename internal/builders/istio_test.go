package builders

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
)

var _ = Describe("Builder for", func() {

	path := "/headers"
	methods := []gatewayv1beta1.HttpMethod{http.MethodGet, http.MethodPost}

	Describe("AuthorizationPolicy", func() {
		It("should build an AuthorizationPolicy", func() {
			name := "testName"
			namespace := "testNs"

			testMatchLabelsKey := "app"
			testMatchLabelsValue := "httpbin"
			testMatchLabels := map[string]string{testMatchLabelsKey: testMatchLabelsValue}
			testScopeA := "scope-a"
			testScopeB := "scope-b"
			testAuthorization := gatewayv1beta1.JwtAuthorization{RequiredScopes: []string{testScopeA, testScopeB}}
			testExpectedScopeKeys := []string{"request.auth.claims[scp]"}
			testRaw := runtime.RawExtension{
				Raw: []byte(`{"authentications": [{"issuer": "testIssuer", "jwksUri": "testJwksUri"}], "authorizations": [{"requiredScopes": ["test"]}]}`),
			}
			testHandler := gatewayv1beta1.Handler{Name: "jwt", Config: &testRaw}
			testAuthenticator := gatewayv1beta1.Authenticator{Handler: &testHandler}
			testAccessStrategies := []*gatewayv1beta1.Authenticator{&testAuthenticator}
			testRequestPrincipal := "testIssuer/*"

			ap := NewAuthorizationPolicyBuilder().WithGenerateName(name).WithNamespace(namespace).
				WithSpec(NewAuthorizationPolicySpecBuilder().
					WithSelector(NewSelectorBuilder().
						WithMatchLabels(testMatchLabelsKey, testMatchLabelsValue).Get()).
					WithRule(NewRuleBuilder().
						WithFrom(NewFromBuilder().
							WithForcedJWTAuthorization(testAccessStrategies).Get()).
						WithTo(NewToBuilder().
							WithOperation(NewOperationBuilder().
								WithPath(path).
								WithMethods(methods).Get()).Get()).
						WithWhenCondition(NewConditionBuilder().
							WithKey("request.auth.claims[scp]").WithValues(testAuthorization.RequiredScopes).Get()).Get()).Get()).
				Get()

			Expect(ap.Name).To(BeEmpty())
			Expect(ap.GenerateName).To(Equal(name))
			Expect(ap.Namespace).To(Equal(namespace))
			Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(testMatchLabels))
			Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(testRequestPrincipal))
			Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).To(Equal(path))
			Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(BeEquivalentTo([]string{http.MethodGet, http.MethodPost}))
			Expect(ap.Spec.Rules[0].When).To(HaveLen(1))
			Expect(ap.Spec.Rules[0].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap.Spec.Rules[0].When[0].Values).To(HaveLen(2))
			Expect(ap.Spec.Rules[0].When[0].Values).To(ContainElements(testScopeA, testScopeB))
		})
	})

	Describe("RequestAuthentication", func() {
		name := "testName"
		namespace := "testNs"

		testMatchLabelsKey := "app"
		testMatchLabelsValue := "httpbin"
		testMatchLabels := map[string]string{testMatchLabelsKey: testMatchLabelsValue}

		It("should build a RequestAuthentication", func() {
			testRaw := runtime.RawExtension{Raw: []byte(`{"authentications": [{"issuer": "testIssuer", "jwksUri": "testJwksUri"}]}`)}
			testHandler := gatewayv1beta1.Handler{Config: &testRaw}
			testAuthenticator := gatewayv1beta1.Authenticator{Handler: &testHandler}
			testAccessStrategies := []*gatewayv1beta1.Authenticator{&testAuthenticator}

			ap := NewRequestAuthenticationBuilder().WithGenerateName(name).WithNamespace(namespace).
				WithSpec(NewRequestAuthenticationSpecBuilder().
					WithSelector(NewSelectorBuilder().WithMatchLabels(testMatchLabelsKey, testMatchLabelsValue).Get()).
					WithJwtRules(*NewJwtRuleBuilder().From(testAccessStrategies).Get()).
					Get()).
				Get()

			Expect(ap.Name).To(BeEmpty())
			Expect(ap.GenerateName).To(Equal(name))
			Expect(ap.Namespace).To(Equal(namespace))
			Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(testMatchLabels))
			Expect(ap.Spec.JwtRules).To(HaveLen(1))
			Expect(ap.Spec.JwtRules[0].Issuer).To(Equal("testIssuer"))
			Expect(ap.Spec.JwtRules[0].JwksUri).To(Equal("testJwksUri"))
			Expect(ap.Spec.JwtRules[0].FromHeaders).To(BeEmpty())
			Expect(ap.Spec.JwtRules[0].FromParams).To(BeEmpty())
			Expect(ap.Spec.JwtRules[0].ForwardOriginalToken).To(BeTrue())
		})

		It("should build an RequestAuthentication with 2 JwtRules", func() {
			testRaw := runtime.RawExtension{
				Raw: []byte(`{"authentications": [{"issuer": "testIssuer1", "jwksUri": "testJwksUri1"}, {"issuer": "testIssuer2", "jwksUri": "testJwksUri2"}]}`),
			}
			testHandler := gatewayv1beta1.Handler{Config: &testRaw}
			testAuthenticator := gatewayv1beta1.Authenticator{Handler: &testHandler}
			testAccessStrategies := []*gatewayv1beta1.Authenticator{&testAuthenticator}

			ap := NewRequestAuthenticationBuilder().WithGenerateName(name).WithNamespace(namespace).
				WithSpec(NewRequestAuthenticationSpecBuilder().
					WithSelector(NewSelectorBuilder().WithMatchLabels(testMatchLabelsKey, testMatchLabelsValue).Get()).
					WithJwtRules(*NewJwtRuleBuilder().From(testAccessStrategies).Get()).
					Get()).
				Get()

			Expect(ap.Spec.JwtRules).To(HaveLen(2))
			Expect(ap.Spec.JwtRules[0].Issuer).To(Equal("testIssuer1"))
			Expect(ap.Spec.JwtRules[0].JwksUri).To(Equal("testJwksUri1"))
			Expect(ap.Spec.JwtRules[0].FromHeaders).To(BeEmpty())
			Expect(ap.Spec.JwtRules[0].FromParams).To(BeEmpty())
			Expect(ap.Spec.JwtRules[0].ForwardOriginalToken).To(BeTrue())
			Expect(ap.Spec.JwtRules[1].Issuer).To(Equal("testIssuer2"))
			Expect(ap.Spec.JwtRules[1].JwksUri).To(Equal("testJwksUri2"))
			Expect(ap.Spec.JwtRules[1].FromHeaders).To(BeEmpty())
			Expect(ap.Spec.JwtRules[1].FromParams).To(BeEmpty())
			Expect(ap.Spec.JwtRules[1].ForwardOriginalToken).To(BeTrue())

		})

		It("should build an RequestAuthentication with fromHeaders", func() {
			testRaw := runtime.RawExtension{
				Raw: []byte(
					`{"authentications": [{"issuer": "testIssuer", "jwksUri": "testJwksUri", "fromHeaders": [{"Name": "testHeader1"}, {"Name": "testHeader2", "Prefix": "testPrefix2"}]}]}`,
				),
			}
			testHandler := gatewayv1beta1.Handler{Config: &testRaw}
			testAuthenticator := gatewayv1beta1.Authenticator{Handler: &testHandler}
			testAccessStrategies := []*gatewayv1beta1.Authenticator{&testAuthenticator}

			ap := NewRequestAuthenticationBuilder().WithGenerateName(name).WithNamespace(namespace).
				WithSpec(NewRequestAuthenticationSpecBuilder().
					WithSelector(NewSelectorBuilder().WithMatchLabels(testMatchLabelsKey, testMatchLabelsValue).Get()).
					WithJwtRules(*NewJwtRuleBuilder().From(testAccessStrategies).Get()).
					Get()).
				Get()

			Expect(ap.Spec.JwtRules).To(HaveLen(1))
			Expect(ap.Spec.JwtRules[0].FromHeaders).To(HaveLen(2))
			Expect(ap.Spec.JwtRules[0].FromHeaders[0].Name).To(Equal("testHeader1"))
			Expect(ap.Spec.JwtRules[0].FromHeaders[0].Prefix).To(BeEmpty())
			Expect(ap.Spec.JwtRules[0].FromHeaders[1].Name).To(Equal("testHeader2"))
			Expect(ap.Spec.JwtRules[0].FromHeaders[1].Prefix).To(Equal("testPrefix2"))
			Expect(ap.Spec.JwtRules[0].FromParams).To(BeEmpty())
			Expect(ap.Spec.JwtRules[0].ForwardOriginalToken).To(BeTrue())
		})

		It("should build an RequestAuthentication with fromParams", func() {
			testRaw := runtime.RawExtension{Raw: []byte(`{"authentications": [{"issuer": "testIssuer", "jwksUri": "testJwksUri", "fromParams": ["param1", "param2"]}]}`)}
			testHandler := gatewayv1beta1.Handler{Config: &testRaw}
			testAuthenticator := gatewayv1beta1.Authenticator{Handler: &testHandler}
			testAccessStrategies := []*gatewayv1beta1.Authenticator{&testAuthenticator}

			ap := NewRequestAuthenticationBuilder().WithGenerateName(name).WithNamespace(namespace).
				WithSpec(NewRequestAuthenticationSpecBuilder().
					WithSelector(NewSelectorBuilder().WithMatchLabels(testMatchLabelsKey, testMatchLabelsValue).Get()).
					WithJwtRules(*NewJwtRuleBuilder().From(testAccessStrategies).Get()).
					Get()).
				Get()

			Expect(ap.Spec.JwtRules).To(HaveLen(1))
			Expect(ap.Spec.JwtRules[0].FromParams).To(HaveLen(2))
			Expect(ap.Spec.JwtRules[0].FromParams[0]).To(Equal("param1"))
			Expect(ap.Spec.JwtRules[0].FromParams[1]).To(Equal("param2"))
			Expect(ap.Spec.JwtRules[0].FromHeaders).To(BeEmpty())
			Expect(ap.Spec.JwtRules[0].ForwardOriginalToken).To(BeTrue())
		})
	})
})
