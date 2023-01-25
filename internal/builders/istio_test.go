package builders

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Builder for", func() {

	path := "/headers"
	methods := []string{"GET", "POST"}

	Describe("AuthorizationPolicy", func() {
		It("should build the object", func() {
			name := "testName"
			namespace := "testNs"

			testMatchLabelsKey := "app"
			testMatchLabelsValue := "httpbin"
			testMatchLabels := map[string]string{testMatchLabelsKey: testMatchLabelsValue}
			testRulesSourceRequestPrincipals := "*"
			testScopeA := "scope-a"
			testScopeB := "scope-b"
			testAuthorization := gatewayv1beta1.JwtAuthorization{RequiredScopes: []string{testScopeA, testScopeB}}
			testExpectedScopeKeys := []string{"request.auth.claims[scp]"}

			ap := AuthorizationPolicyBuilder().GenerateName(name).Namespace(namespace).
				Spec(AuthorizationPolicySpecBuilder().
					Selector(SelectorBuilder().
						MatchLabels(testMatchLabelsKey, testMatchLabelsValue)).
					Rule(RuleBuilder().
						RuleFrom(RuleFromBuilder().
							Source()).
						RuleTo(RuleToBuilder().
							Operation(OperationBuilder().
								Path(path).
								Methods(methods))).
						RuleCondition(RuleConditionBuilder().
							From("request.auth.claims[scp]", &testAuthorization)))).
				Get()

			Expect(ap.Name).To(BeEmpty())
			Expect(ap.GenerateName).To(Equal(name))
			Expect(ap.Namespace).To(Equal(namespace))
			Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(testMatchLabels))
			Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(testRulesSourceRequestPrincipals))
			Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).To(Equal(path))
			Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(BeEquivalentTo(methods))
			Expect(ap.Spec.Rules[0].When).To(HaveLen(2))
			Expect(ap.Spec.Rules[0].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
			Expect(ap.Spec.Rules[0].When[0].Values).To(HaveLen(1))
			Expect(ap.Spec.Rules[0].When[1].Values).To(HaveLen(1))
			Expect(ap.Spec.Rules[0].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
			if ap.Spec.Rules[0].When[0].Values[0] == testScopeA {
				Expect(ap.Spec.Rules[0].When[1].Values[0]).To(Equal(testScopeB))
			} else {
				Expect(ap.Spec.Rules[0].When[1].Values[0]).To(Equal(testScopeA))
				Expect(ap.Spec.Rules[0].When[0].Values[0]).To(Equal(testScopeB))
			}
		})
	})

	Describe("RequestAuthentication", func() {
		It("should build the object", func() {
			name := "testName"
			namespace := "testNs"

			testMatchLabelsKey := "app"
			testMatchLabelsValue := "httpbin"
			testMatchLabels := map[string]string{testMatchLabelsKey: testMatchLabelsValue}
			testRaw := runtime.RawExtension{Raw: []byte(`{"authentications": [{"issuer": "testIssuer", "jwksUri": "testJwksUri"}]}`)}
			testHandler := gatewayv1beta1.Handler{Config: &testRaw}
			testAuthenticator := gatewayv1beta1.Authenticator{Handler: &testHandler}
			testAccessStrategies := []*gatewayv1beta1.Authenticator{&testAuthenticator}

			ap := RequestAuthenticationBuilder().GenerateName(name).Namespace(namespace).
				Spec(RequestAuthenticationSpecBuilder().
					Selector(SelectorBuilder().
						MatchLabels(testMatchLabelsKey, testMatchLabelsValue)).
					JwtRules(JwtRuleBuilder().From(testAccessStrategies))).
				Get()

			Expect(ap.Name).To(BeEmpty())
			Expect(ap.GenerateName).To(Equal(name))
			Expect(ap.Namespace).To(Equal(namespace))
			Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(testMatchLabels))
			Expect(ap.Spec.JwtRules[0].Issuer).To(Equal("testIssuer"))
			Expect(ap.Spec.JwtRules[0].JwksUri).To(Equal("testJwksUri"))
		})
	})
})
