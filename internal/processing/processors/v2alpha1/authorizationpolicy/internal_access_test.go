package authorizationpolicy_test

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/authorizationpolicy"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("Internal access from cluster APs", func() {
	It("should create an AP for internal access, one if only one service is used", func() {
		// given
		ruleJwt := newNoAuthRuleBuilderWithDummyData().
			withPath("/abc").
			build()

		noAuthRule := newNoAuthRuleBuilderWithDummyData().
			withPath("/def").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt, noAuthRule).
			build()
		svc := newServiceBuilderWithDummyData().build()
		client := getFakeClient(svc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(3))

		goodAps := 0
		for i := 0; i < 3; i++ {
			ap := result[i].Obj.(*securityv1beta1.AuthorizationPolicy)
			if len(ap.Spec.Rules[0].To) == 0 {
				Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
				Expect(ap.Spec.Rules[0].From[0].Source.NotPrincipals).To(ConsistOf("cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"))
				goodAps++
			} else if len(ap.Spec.Rules[0].To) == 1 {
				if ap.Spec.Rules[0].To[0].Operation.Paths[0] == "/abc" {
					goodAps++
				} else if ap.Spec.Rules[0].To[0].Operation.Paths[0] == "/def" {
					goodAps++
				}
			}
		}
		Expect(goodAps).To(Equal(3))
	})

	It("should create an AP for internal access, two if two services are used", func() {
		// given
		ruleJwt := newNoAuthRuleBuilderWithDummyData().
			withPath("/abc").
			build()

		noAuthRule := newNoAuthRuleBuilderWithDummyData().
			withPath("/def").
			withServiceName("different-service").
			build()

		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(ruleJwt, noAuthRule).
			build()
		svc := newServiceBuilderWithDummyData().build()
		differentSvc := newServiceBuilderWithDummyData().withName("different-service").addSelector("a", "b").build()
		client := getFakeClient(svc, differentSvc)
		processor := authorizationpolicy.NewProcessor(&testLogger, apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(4))

		for _, apResult := range result {
			ap := apResult.Obj.(*securityv1beta1.AuthorizationPolicy)

			if len(ap.Spec.Rules[0].To) > 0 {
				Expect(len(ap.Spec.Rules[0].To[0].Operation.Paths)).To(Equal(1))
				if ap.Spec.Rules[0].To[0].Operation.Paths[0] == "/abc" {
					Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElement("/abc"))
				} else if ap.Spec.Rules[0].To[0].Operation.Paths[0] == "/def" {
					Expect(ap.Spec.Rules[0].To[0].Operation.Paths).To(ContainElement("/def"))
				}
			} else {
				Expect(len(ap.Spec.Rules[0].From)).To(Equal(1))
				Expect(ap.Spec.Rules[0].From[0].Source.NotPrincipals).To(ConsistOf("cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"))
			}
		}
	})
})
