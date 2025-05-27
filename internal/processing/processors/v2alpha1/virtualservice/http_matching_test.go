package virtualservice_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/kyma-project/api-gateway/internal/builders/builders_test/v2alpha1_test"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
)

var _ = Describe("HTTP matching", func() {
	var client client.Client
	var processor processors.Processor
	BeforeEach(func() {
		client = GetFakeClient()
	})
	var _ = DescribeTable("Different methods on same path",
		func(apiRule *gatewayv2alpha1.APIRule, verifiers []verifier, expectedError error, expectedActions ...string) {
			processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule, nil)
			checkVirtualServices(client, processor, verifiers, expectedError, expectedActions...)
		},
		Entry("from two rules with different methods on the same path should create two HTTP routes with different methods",
			NewAPIRuleBuilderWithDummyData().
				WithRules(NewRuleBuilder().WithMethods(http.MethodGet).WithPath("/").NoAuth().Build(),
					NewRuleBuilder().WithMethods(http.MethodPut).WithPath("/").WithJWTAuthn("example.com", "https://jwks.example.com", nil, nil).Build()).Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http).To(HaveLen(2))

					Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET)$"))
					Expect(vs.Spec.Http[1].Match[0].Method.GetRegex()).To(Equal("^(PUT)$"))
				},
			}, nil, "create"),

		Entry("from one rule with two methods on the same path should create one HTTP route with regex matching both methods",
			NewAPIRuleBuilderWithDummyData().
				WithRules(NewRuleBuilder().WithMethods(http.MethodGet, http.MethodPut).WithPath("/").NoAuth().Build()).
				Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http).To(HaveLen(1))

					Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET|PUT)$"))
				},
			}, nil, "create"),
	)
})
