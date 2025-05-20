package requestauthentication_test

import (
	"context"

	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/requestauthentication"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
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
})
