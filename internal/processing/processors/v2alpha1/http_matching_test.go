package v2alpha1_test

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP matching", func() {
	var client client.Client
	var processor processors.VirtualServiceProcessor
	BeforeEach(func() {
		client = GetFakeClient()
	})
	var _ = DescribeTable("Different methods on same path",
		func(apiRule *gatewayv2alpha1.APIRule, verifiers []verifier, expectedActions ...string) {
			processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule)
			checkVirtualServices(client, processor, verifiers, expectedActions...)
		},
		Entry("from two rules with different methods on the same path should create two HTTP routes with different methods",
			newAPIRuleBuilderWithDummyData().
				WithRules(newRuleBuilder().WithMethods(http.MethodGet).WithPath("/").NoAuth().Build(),
					newRuleBuilder().WithMethods(http.MethodPut).WithPath("/").WithJWTAuthn("example.com", "https://jwks.example.com", nil, nil).Build()).Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http).To(HaveLen(2))

					Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET)$"))
					Expect(vs.Spec.Http[1].Match[0].Method.GetRegex()).To(Equal("^(PUT)$"))
				},
			}, "create"),

		Entry("from one rule with two methods on the same path should create one HTTP route with regex matching both methods",
			newAPIRuleBuilderWithDummyData().
				WithRules(newRuleBuilder().WithMethods(http.MethodGet, http.MethodPut).WithPath("/").NoAuth().Build()).
				Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http).To(HaveLen(1))

					Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET|PUT)$"))
				},
			}, "create"),
	)
})
