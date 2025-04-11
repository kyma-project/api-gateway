package authorizationpolicy_test

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/slices"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("Processing NoAuth rules", func() {
	requiredScopeA := "scope-a"
	requiredScopeB := "scope-b"

	It("should create AP for noAuth with From spec having Source.Principals == cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account", func() {
		// given
		headersPath := "/headers"
		imgPath := "/img"
		ruleJwt := newJwtRuleBuilderWithDummyData().
			withPath(imgPath).
			addJwtAuthorizationRequiredScopes(requiredScopeA, requiredScopeB).
			build()

		ruleNoAuth := newNoAuthRuleBuilderWithDummyData().
			withPath(headersPath).
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleNoAuth, ruleJwt).
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

			expectedHandlers := []string{headersPath, imgPath}
			Expect(slices.Contains(expectedHandlers, ap.Spec.Rules[0].To[0].Operation.Paths[0])).To(BeTrue())

			switch ap.Spec.Rules[0].To[0].Operation.Paths[0] {
			case headersPath:
				Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
				Expect(ap.Spec.Rules[0].From[0].Source.Principals[0]).To(Equal("cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"))
			case imgPath:
				Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
			}

			Expect(len(ap.Spec.Rules)).To(BeElementOf([]int{1, 3}))
			if len(ap.Spec.Rules) == 3 {
				for i := 0; i < 3; i++ {
					Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
					Expect(ap.Spec.Rules[i].When).To(HaveLen(2))
					Expect(ap.Spec.Rules[i].When[0].Key).To(BeElementOf(testExpectedScopeKeys))
					Expect(ap.Spec.Rules[i].When[0].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
					Expect(ap.Spec.Rules[i].When[1].Key).To(BeElementOf(testExpectedScopeKeys))
					Expect(ap.Spec.Rules[i].When[1].Values[0]).To(BeElementOf(requiredScopeA, requiredScopeB))
				}
			} else {
				Expect(len(ap.Spec.Rules)).To(Equal(1))
			}
		}
	})

})
