package virtualservice_test

import (
	"context"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders"
	. "github.com/kyma-project/api-gateway/internal/builders/builders_test/v2alpha1_test"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	istioapiv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ObjectChange", func() {
	It("should return create action when there is no VirtualService on cluster", func() {
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

	It("should return update action when there is a matching VirtualService on cluster", func() {
		// given
		apiRuleBuilder := NewAPIRuleBuilderWithDummyData()
		processor := processors.NewVirtualServiceProcessor(GetTestConfig(), apiRuleBuilder.Build(), nil)
		result, err := processor.EvaluateReconciliation(context.Background(), GetFakeClient())

		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))

		// when
		processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRuleBuilder.WithHosts("newHost.com").Build(), nil)
		result, err = processor.EvaluateReconciliation(context.Background(), GetFakeClient(result[0].Obj.(*networkingv1beta1.VirtualService)))

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("update"))
	})
})

var _ = Describe("Fully configured APIRule happy path", func() {
	It("should create a VirtualService with all the configured values", func() {
		apiRule := NewAPIRuleBuilder().
			WithGateway("example/example").
			WithHosts("example.com", "goat.com").
			WithService("example-service", "example-namespace", 8080).
			WithTimeout(180).
			WithCORSPolicy(NewCorsPolicyBuilder().
				WithAllowOrigins([]map[string]string{{"exact": "example.com"}}).
				WithAllowMethods([]string{"GET", "POST"}).
				WithAllowHeaders([]string{"header1", "header2"}).
				WithExposeHeaders([]string{"header3", "header4"}).
				WithAllowCredentials(true).
				WithMaxAge(600).
				Build()).
			WithRules(
				NewRuleBuilder().
					WithService("another-service", "another-namespace", 9999).
					WithMethods("GET", "POST").
					WithPath("/").
					WithJWTAuthn("example.com", "https://jwks.example.com", nil, nil).
					WithTimeout(10).Build(),

				NewRuleBuilder().
					WithMethods("PUT").
					WithPath("/*").
					WithRequest(NewRequestModifier().
						WithHeaders(map[string]string{"header1": "value1"}).
						WithCookies(map[string]string{"cookie1": "value1"}).
						Build()).
					NoAuth().Build(),
			).
			Build()

		client := GetFakeClient()
		processor := processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule, nil)
		checkVirtualServices(client, processor, []verifier{
			func(vs *networkingv1beta1.VirtualService) {
				Expect(vs.Spec.Hosts).To(ConsistOf("example.com", "goat.com"))
				Expect(vs.Spec.Gateways).To(ConsistOf("example/example"))
				Expect(vs.Spec.Http).To(HaveLen(2))

				Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET|POST)$"))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal("/"))
				Expect(vs.Spec.Http[0].Route[0].Destination.Host).To(Equal("another-service.another-namespace.svc.cluster.local"))
				Expect(vs.Spec.Http[0].Route[0].Destination.Port.Number).To(Equal(uint32(9999)))

				Expect(vs.Spec.Http[0].CorsPolicy).NotTo(BeNil())
				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins).To(HaveLen(1))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowOrigins[0]).To(Equal(&istioapiv1beta1.StringMatch{MatchType: &istioapiv1beta1.StringMatch_Exact{Exact: "example.com"}}))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowMethods).To(ConsistOf("GET", "POST"))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowHeaders).To(ConsistOf("header1", "header2"))
				Expect(vs.Spec.Http[0].CorsPolicy.ExposeHeaders).To(ConsistOf("header3", "header4"))
				Expect(vs.Spec.Http[0].CorsPolicy.AllowCredentials.GetValue()).To(BeTrue())
				Expect(vs.Spec.Http[0].CorsPolicy.MaxAge.Seconds).To(Equal(int64(600)))

				Expect(vs.Spec.Http[0].Timeout.Seconds).To(Equal(int64(10)))

				Expect(vs.Spec.Http[1].Match[0].Method.GetRegex()).To(Equal("^(PUT)$"))
				Expect(vs.Spec.Http[1].Match[0].Uri.GetPrefix()).To(Equal("/"))
				Expect(vs.Spec.Http[1].Route[0].Destination.Host).To(Equal("example-service.example-namespace.svc.cluster.local"))
				Expect(vs.Spec.Http[1].Route[0].Destination.Port.Number).To(Equal(uint32(8080)))

				Expect(vs.Spec.Http[1].CorsPolicy).NotTo(BeNil())
				Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins).To(HaveLen(1))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowOrigins[0]).To(Equal(&istioapiv1beta1.StringMatch{MatchType: &istioapiv1beta1.StringMatch_Exact{Exact: "example.com"}}))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowMethods).To(ConsistOf("GET", "POST"))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowHeaders).To(ConsistOf("header1", "header2"))
				Expect(vs.Spec.Http[1].CorsPolicy.ExposeHeaders).To(ConsistOf("header3", "header4"))
				Expect(vs.Spec.Http[1].CorsPolicy.AllowCredentials.GetValue()).To(BeTrue())
				Expect(vs.Spec.Http[1].CorsPolicy.MaxAge.Seconds).To(Equal(int64(600)))

				Expect(vs.Spec.Http[1].Headers).NotTo(BeNil())
				Expect(vs.Spec.Http[1].Headers.Request).NotTo(BeNil())
				Expect(vs.Spec.Http[1].Headers.Request.Set).To(HaveKeyWithValue("x-forwarded-host", "example.com"))
				Expect(vs.Spec.Http[1].Headers.Request.Set).To(HaveKeyWithValue("header1", "value1"))
				Expect(vs.Spec.Http[1].Headers.Request.Set).To(HaveKeyWithValue("Cookie", "cookie1=value1"))

				Expect(vs.Spec.Http[1].Timeout.Seconds).To(Equal(int64(180)))
			},
		}, nil, "create")

	})
})

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

	It("should create a VirtualService with '/' prefix, when rule in APIRule applies to all paths", func() {
		apiRule := NewAPIRuleBuilder().
			WithGateway("example/example").
			WithHosts("example.com").
			WithService("example-service", "example-namespace", 8080).
			WithTimeout(180).
			WithRules(
				NewRuleBuilder().
					WithMethods("GET").
					WithPath("/*").
					NoAuth().Build(),
			).
			Build()

		client := GetFakeClient()
		processor := processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule, nil)

		checkVirtualServices(client, processor, []verifier{
			func(vs *networkingv1beta1.VirtualService) {
				Expect(vs.Spec.Hosts).To(ConsistOf("example.com"))
				Expect(vs.Spec.Gateways).To(ConsistOf("example/example"))
				Expect(vs.Spec.Http).To(HaveLen(1))

				Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET)$"))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetPrefix()).To(Equal("/"))
			},
		}, nil, "create")
	})

	It("should create a VirtualService with regex string match for path, when rule in APIRule spcifify sub-paths with '*'", func() {
		apiRule := NewAPIRuleBuilder().
			WithGateway("example/example").
			WithHosts("example.com").
			WithService("example-service", "example-namespace", 8080).
			WithTimeout(180).
			WithRules(
				NewRuleBuilder().
					WithMethods("GET").
					WithPath("/callback/*").
					NoAuth().Build(),
			).
			Build()

		client := GetFakeClient()
		processor := processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule, nil)

		checkVirtualServices(client, processor, []verifier{
			func(vs *networkingv1beta1.VirtualService) {
				Expect(vs.Spec.Hosts).To(ConsistOf("example.com"))
				Expect(vs.Spec.Gateways).To(ConsistOf("example/example"))
				Expect(vs.Spec.Http).To(HaveLen(1))

				Expect(vs.Spec.Http[0].Match[0].Method.GetRegex()).To(Equal("^(GET)$"))
				Expect(vs.Spec.Http[0].Match[0].Uri.GetRegex()).To(Equal("/callback/.*"))
			},
		}, nil, "create")
	})
})

func checkVirtualServices(c client.Client, processor processors.VirtualServiceProcessor, verifiers []verifier, expectedError error, expectedActions ...string) {
	result, err := processor.EvaluateReconciliation(context.Background(), c)
	if expectedError != nil {
		Expect(result).To(HaveLen(len(expectedActions)))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(expectedError.Error()))
		return
	}

	Expect(err).ToNot(HaveOccurred())
	Expect(result).To(HaveLen(len(expectedActions)))
	for i, action := range expectedActions {
		Expect(result[i].Action.String()).To(Equal(action))
	}

	for i, v := range verifiers {
		v(result[i].Obj.(*networkingv1beta1.VirtualService))
	}
}

type verifier func(*networkingv1beta1.VirtualService)

type mockVirtualServiceCreator struct{}

func (r mockVirtualServiceCreator) Create(_ *gatewayv2alpha1.APIRule) (*networkingv1beta1.VirtualService, error) {
	return builders.VirtualService().Get(), nil
}
