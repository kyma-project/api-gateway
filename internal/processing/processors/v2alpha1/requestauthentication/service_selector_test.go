package requestauthentication_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/requestauthentication"
)

var _ = Describe("Service has custom selector", func() {

	It("should create RA with selector from service", func() {
		// given: New resources
		jwtRule := newJwtRuleBuilderWithDummyData().build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(jwtRule).
			build()
		svc := newServiceBuilder().
			withName("example-service").
			withNamespace("example-namespace").
			addSelector("custom", "example-service").
			build()
		client := getFakeClient(svc)

		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels).To(HaveLen(1))
		Expect(ra.Spec.Selector.MatchLabels["custom"]).To(Equal(serviceName))
	})

	It("should create RA with selector from service in different namespace", func() {
		// given: New resources
		differentNamespace := "different-namespace"

		jwtRule := newJwtRuleBuilderWithDummyData().
			withServiceNamespace(differentNamespace).
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(jwtRule).
			build()
		svc := newServiceBuilder().
			withName("example-service").
			withNamespace(differentNamespace).
			addSelector("custom", "example-service").
			build()

		client := getFakeClient(svc)

		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels).To(HaveLen(1))
		Expect(ra.Spec.Selector.MatchLabels["custom"]).To(Equal(serviceName))
	})

	It("should create RA with selector from service with multiple selector labels", func() {
		// given: New resources

		jwtRule := newJwtRuleBuilderWithDummyData().
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(jwtRule).
			build()
		svc := newServiceBuilder().
			withName("example-service").
			withNamespace("example-namespace").
			addSelector("custom", "example-service").
			addSelector("second-custom", "foo").
			build()

		client := getFakeClient(svc)

		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))

		ra := result[0].Obj.(*securityv1beta1.RequestAuthentication)

		Expect(ra).NotTo(BeNil())
		Expect(ra.Spec.Selector.MatchLabels).To(HaveLen(2))
		Expect(ra.Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", serviceName))
		Expect(ra.Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "foo"))
	})

	It("should create two RA with selectors without app label from service with multiple selector labels", func() {
		// given: New resources

		jwtRule1 := newJwtRuleBuilderWithDummyData().
			withServiceName("example-service-1").
			withPath("/first").
			build()
		jwtRule2 := newJwtRuleBuilderWithDummyData().
			withServiceName("example-service-2").
			withPath("/second").
			build()
		apiRule := newAPIRuleBuilderWithDummyData().
			withRules(jwtRule1, jwtRule2).
			build()
		svc1 := newServiceBuilder().
			withName("example-service-1").
			withNamespace("example-namespace").
			addSelector("custom", "example-service-1").
			addSelector("second-custom", "foo-1").
			build()

		svc2 := newServiceBuilder().
			withName("example-service-2").
			withNamespace("example-namespace").
			addSelector("custom", "example-service-2").
			addSelector("second-custom", "foo-2").
			build()

		client := getFakeClient(svc1, svc2)

		processor := requestauthentication.NewProcessor(apiRule)

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), client)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(2))
		var ra1, ra2 *securityv1beta1.RequestAuthentication
		if result[0].Obj.(*securityv1beta1.RequestAuthentication).Spec.Selector.MatchLabels["custom"] == "example-service-1" {
			ra1 = result[0].Obj.(*securityv1beta1.RequestAuthentication)
			ra2 = result[1].Obj.(*securityv1beta1.RequestAuthentication)
		} else {
			ra1 = result[1].Obj.(*securityv1beta1.RequestAuthentication)
			ra2 = result[0].Obj.(*securityv1beta1.RequestAuthentication)
		}

		Expect(ra1).NotTo(BeNil())
		Expect(ra1.Spec.Selector.MatchLabels).To(HaveLen(2))
		Expect(ra1.Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", "example-service-1"))
		Expect(ra1.Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "foo-1"))

		Expect(ra2).NotTo(BeNil())
		Expect(ra2.Spec.Selector.MatchLabels).To(HaveLen(2))
		Expect(ra2.Spec.Selector.MatchLabels).To(HaveKeyWithValue("custom", "example-service-2"))
		Expect(ra2.Spec.Selector.MatchLabels).To(HaveKeyWithValue("second-custom", "foo-2"))
	})
})
