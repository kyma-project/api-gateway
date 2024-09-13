package authorizationpolicy_test

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"

	"github.com/kyma-project/api-gateway/internal/builders"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("JwtAuthorization Policy Processor", func() {
	testAudience1 := "kyma-goats"
	testAudience2 := "kyma-goats2"

	It("should produce one AP for a rule with two audiences", func() {
		// given
		rule := newJwtRuleBuilderWithDummyData().
			addJwtAuthorizationAudiences(testAudience1, testAudience2).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(rule).
			build()

		svc := newServiceBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, false)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.Spec.Rules).To(HaveLen(1))
		Expect(ap1.Spec.Rules[0].When).To(HaveLen(2))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.claims[aud]").WithValues([]string{testAudience1}).Get()))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.claims[aud]").WithValues([]string{testAudience2}).Get()))
	})

	It("should produce one AP for a rule with two scopes and two audiences", func() {
		// given
		rule := newJwtRuleBuilderWithDummyData().
			addJwtAuthorization([]string{"scope-a", "scope-b"}, []string{testAudience1, testAudience2}).
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(rule).
			build()
		svc := newServiceBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule, false)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ap1 := result[0].Obj.(*securityv1beta1.AuthorizationPolicy)

		Expect(ap1).NotTo(BeNil())
		Expect(ap1.Spec.Rules).To(HaveLen(3))
		Expect(ap1.Spec.Rules[0].When).To(HaveLen(4))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.claims[aud]").WithValues([]string{testAudience1}).Get()))
		Expect(ap1.Spec.Rules[0].When).To(ContainElement(builders.NewConditionBuilder().WithKey("request.auth.claims[aud]").WithValues([]string{testAudience2}).Get()))
	})
})
