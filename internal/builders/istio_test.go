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
								Methods(methods))))).
				Get()

			Expect(ap.Name).To(BeEmpty())
			Expect(ap.GenerateName).To(Equal(name))
			Expect(ap.Namespace).To(Equal(namespace))
			Expect(ap.Spec.Selector.MatchLabels).To(BeEquivalentTo(testMatchLabels))
			Expect(ap.Spec.Rules[0].From[0].Source.RequestPrincipals[0]).To(Equal(testRulesSourceRequestPrincipals))
			Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).To(Equal(path))
			Expect(ap.Spec.Rules[0].To[0].Operation.Methods).To(BeEquivalentTo(methods))
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
