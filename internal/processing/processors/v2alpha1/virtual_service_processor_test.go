package v2alpha1_test

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	istioapiv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VirtualServiceProcessor", func() {
	It("should create virtual service when no virtual service exists", func() {
		// given
		processor := processors.VirtualServiceProcessor{
			ApiRule: &gatewayv2alpha1.APIRule{},
			Creator: mockVirtualServiceCreator{},
		}

		// when
		result, err := processor.EvaluateReconciliation(context.Background(), GetFakeClient())

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})
})

func checkVirtualServices(c client.Client, processor processors.VirtualServiceProcessor, verifiers []verifier, expectedActions ...string) {
	result, err := processor.EvaluateReconciliation(context.Background(), c)
	Expect(err).To(BeNil())
	Expect(result).To(HaveLen(len(expectedActions)))
	for i, action := range expectedActions {
		Expect(result[i].Action.String()).To(Equal(action))
	}

	for i, v := range verifiers {
		v(result[i].Obj.(*networkingv1beta1.VirtualService))
	}
}

type verifier func(*networkingv1beta1.VirtualService)

var _ = Describe("Hosts", func() {
	var client client.Client
	var processor processors.VirtualServiceProcessor
	BeforeEach(func() {
		client = GetFakeClient()
	})

	DescribeTable("Hosts",
		func(apiRule *gatewayv2alpha1.APIRule, verifiers []verifier, expectedActions ...string) {
			processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule)
			checkVirtualServices(client, processor, verifiers, expectedActions...)
		},

		Entry("should set the host correctly",
			newAPIRuleBuilder().WithGateway("example/example").WithHost("example.com").Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Hosts).To(ConsistOf("example.com"))
				},
			}, "create"),

		Entry("should set multiple hosts correctly",
			newAPIRuleBuilder().WithGateway("example/example").WithHosts("example.com", "goat.com").Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Hosts).To(ConsistOf("example.com", "goat.com"))
				},
			}, "create"),
	)
})

var _ = Describe("CORS", func() {
	var client client.Client
	var processor processors.VirtualServiceProcessor
	BeforeEach(func() {
		client = GetFakeClient()
	})

	DescribeTable("CORS",
		func(apiRule *gatewayv2alpha1.APIRule, verifiers []verifier, expectedActions ...string) {
			processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule)
			checkVirtualServices(client, processor, verifiers, expectedActions...)
		},

		Entry("should set default empty values in VirtualService CORSPolicy when no CORS configuration is set in APIRule",
			newAPIRuleBuilderWithDummyData().Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http[0].CorsPolicy).To(BeNil())

					Expect(vs.Spec.Http[0].Headers.Response.Remove).To(ConsistOf([]string{
						builders.ExposeHeadersName,
						builders.MaxAgeName,
						builders.AllowHeadersName,
						builders.AllowCredentialsName,
						builders.AllowMethodsName,
						builders.AllowOriginName,
					}))
				},
			}, "create"),

		Entry("should apply all CORSPolicy headers correctly",
			newAPIRuleBuilderWithDummyData().WithCORSPolicy(
				newCorsPolicyBuilder().
					WithAllowOrigins([]map[string]string{{"exact": "example.com"}}).
					WithAllowMethods([]string{"GET", "POST"}).
					WithAllowHeaders([]string{"header1", "header2"}).
					WithExposeHeaders([]string{"header3", "header4"}).
					WithAllowCredentials(true).
					WithMaxAge(600).
					Build()).
				Build(),
			[]verifier{func(vs *networkingv1beta1.VirtualService) {
				Expect(vs.Spec.Http[0].CorsPolicy).NotTo(BeNil())
				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(HaveLen(1))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins[0]).To(Equal(&istioapiv1beta1.StringMatch{MatchType: &istioapiv1beta1.StringMatch_Exact{Exact: "example.com"}}))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(ConsistOf("GET", "POST"))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(ConsistOf("header1", "header2"))
				Expect(vs.Spec.Http[0].CorsPolicy.ExposeHeaders).To(ConsistOf("header3", "header4"))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowCredentials.GetValue()).To(BeTrue())
				Expect(vs.Spec.Http[0].CorsPolicy.MaxAge.Seconds).To(Equal(int64(600)))

				Expect(vs.Spec.Http[0].Headers.Response.Remove).To(ConsistOf([]string{
					builders.ExposeHeadersName,
					builders.MaxAgeName,
					builders.AllowHeadersName,
					builders.AllowCredentialsName,
					builders.AllowMethodsName,
					builders.AllowOriginName,
				}))
			}}, "create"),
	)
})

var _ = Describe("GetVirtualServiceHttpTimeout", func() {
	It("should return default of 180s when no timeout is set", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{}
		rule := gatewayv2alpha1.Rule{}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(180)))
	})

	It("should return the timeout set in the rule when it is set and APIRule has different value", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(20)),
		}
		rule := gatewayv2alpha1.Rule{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(10)),
		}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(10)))
	})

	It("should return the timeout set in the rule when it is set and APIRule timeout is not", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{}
		rule := gatewayv2alpha1.Rule{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(10)),
		}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(10)))
	})

	It("should return the timeout set in the APIRule it is set and rule timeout is not", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(20)),
		}
		rule := gatewayv2alpha1.Rule{}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(10)))
	})
})

type mockVirtualServiceCreator struct{}

func (r mockVirtualServiceCreator) Create(_ *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	return builders.VirtualService().Get(), nil
}
