package authorizationpolicy_test

import (
	"context"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders/builders_test/v2alpha1_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/slices"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("Processing ExtAuth rules", func() {
	audience := "audience"
	extAuthAuthorizer := "test-authorizer"

	It("should create Custom AP and Allow from ingress-gateway for ExtAuth authorizers", func() {
		// given
		headersPath := "/headers"
		ruleExtAuth := v2alpha1_test.NewRuleBuilder().
			WithPath(headersPath).
			WithExtAuth(
				v2alpha1_test.NewExtAuthBuilder().
					WithAuthorizers(extAuthAuthorizer).
					Build()).
			Build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleExtAuth).
			build()

		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		results, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(results).To(HaveLen(2))

		for _, result := range results {
			Expect(result.Obj).To(BeAssignableToTypeOf(&securityv1beta1.AuthorizationPolicy{}))
			ap := result.Obj.(*securityv1beta1.AuthorizationPolicy)
			switch ap.Spec.GetAction() {
			case v1beta1.AuthorizationPolicy_CUSTOM:
				Expect(ap.Spec.Rules).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).To(Equal(headersPath))
				Expect(ap.Spec.Rules[0].To[0].Operation.Hosts).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To[0].Operation.Hosts[0]).To(Equal("example-host.example.com"))

				Expect(ap.Spec.GetProvider().Name).To(Equal(extAuthAuthorizer))
			case v1beta1.AuthorizationPolicy_ALLOW:
				Expect(ap.Spec.Rules).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To[0].Operation.Paths[0]).To(Equal(headersPath))
				Expect(ap.Spec.Rules[0].To[0].Operation.Hosts).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].To[0].Operation.Hosts[0]).To(Equal("example-host.example.com"))
			default:
				Fail("Expected Custom or Allow AuthorizationPolicy")
			}
		}

	})

	It("should create AP for ExtAuth restrictions", func() {
		// given
		headersPath := "/headers"
		ruleJwt := v2alpha1_test.NewRuleBuilder().
			WithPath(headersPath).
			WithExtAuth(
				v2alpha1_test.NewExtAuthBuilder().
					WithAuthorizers(extAuthAuthorizer).
					WithRestrictionAuthorization(&v2alpha1.JwtAuthorization{
						Audiences: []string{audience},
					}).
					Build()).
			Build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt).
			build()
		svc := newServiceBuilderWithDummyData().build()
		gateway := newGatewayBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, gateway)

		// when
		results, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(results).To(HaveLen(2))

		for _, result := range results {
			ap := result.Obj.(*securityv1beta1.AuthorizationPolicy)

			Expect(ap).NotTo(BeNil())
			Expect(len(ap.Spec.Rules[0].To)).To(Equal(1))
			Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))

			expectedHandlers := []string{headersPath}
			Expect(slices.Contains(expectedHandlers, ap.Spec.Rules[0].To[0].Operation.Paths[0])).To(BeTrue())

			switch ap.Spec.Action {
			case v1beta1.AuthorizationPolicy_CUSTOM:
				Expect(ap.Spec.GetProvider().Name).To(Equal(extAuthAuthorizer))
			default:
				Expect(ap.Spec.Action).To(Equal(v1beta1.AuthorizationPolicy_ALLOW))
				Expect(ap.Spec.Rules[0].When).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].When[0].Key).To(Equal("request.auth.claims[aud]"))
				Expect(ap.Spec.Rules[0].When[0].Values).To(HaveLen(1))
				Expect(ap.Spec.Rules[0].When[0].Values[0]).To(Equal(audience))
			}

			Expect(len(ap.Spec.Rules)).To(BeElementOf([]int{1, 3}))
			if len(ap.Spec.Rules) == 3 {
				for i := 0; i < 3; i++ {
					Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
					Expect(ap.Spec.Rules[i].When).To(HaveLen(2))
					Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
					Expect(ap.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
				}
			} else {
				Expect(len(ap.Spec.Rules)).To(Equal(1))
			}
		}
	})

})
