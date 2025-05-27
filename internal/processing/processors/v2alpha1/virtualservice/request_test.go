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

var _ = Describe("Mutators", func() {
	var client client.Client
	var processor processors.VirtualServiceProcessor
	BeforeEach(func() {
		client = GetFakeClient()
	})

	DescribeTable("Mutators",
		func(apiRule *gatewayv2alpha1.APIRule, verifiers []verifier, expectedError error, expectedActions ...string) {
			processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule, nil)
			checkVirtualServices(client, processor, verifiers, expectedError, expectedActions...)
		},

		Entry("should set only x-forwarded-host header when rule does not use any mutators",
			NewAPIRuleBuilderWithDummyDataWithNoAuthRule().Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http[0].Headers).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveLen(1))
					Expect(vs.Spec.Http[0].Headers.Request.Set["x-forwarded-host"]).To(Equal("example-host.example.com"))
				},
			}, nil, "create"),

		Entry("should set Headers on request when rule uses HeadersMutator",
			NewAPIRuleBuilderWithDummyData().
				WithRules(NewRuleBuilder().WithMethods(http.MethodGet).WithPath("/").WithRequest(
					NewRequestModifier().WithHeaders(map[string]string{"header1": "value1"}).Build(),
				).NoAuth().Build()).Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http[0].Headers).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveLen(2))
					Expect(vs.Spec.Http[0].Headers.Request.Set["header1"]).To(Equal("value1"))
					Expect(vs.Spec.Http[0].Headers.Request.Set["x-forwarded-host"]).To(Equal("example-host.example.com"))
				},
			}, nil, "create"),

		Entry("should set Cookie header on request when rule uses CookieMutator",
			NewAPIRuleBuilderWithDummyData().
				WithRules(NewRuleBuilder().WithMethods(http.MethodGet).WithPath("/").WithRequest(
					NewRequestModifier().WithCookies(map[string]string{"header1": "value1"}).Build(),
				).NoAuth().Build()).Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http[0].Headers).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveLen(2))
					Expect(vs.Spec.Http[0].Headers.Request.Set["Cookie"]).To(Equal("header1=value1"))
					Expect(vs.Spec.Http[0].Headers.Request.Set["x-forwarded-host"]).To(Equal("example-host.example.com"))
				},
			}, nil, "create"),

		Entry("should set Cookie header and custom header on request when rule uses CookieMutator and HeadersMutator",
			NewAPIRuleBuilderWithDummyData().
				WithRules(NewRuleBuilder().WithMethods(http.MethodGet).WithPath("/").WithRequest(
					NewRequestModifier().
						WithCookies(map[string]string{"header1": "value1"}).
						WithHeaders(map[string]string{"header2": "value2"}).
						Build(),
				).NoAuth().Build()).Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Http[0].Headers).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request).NotTo(BeNil())
					Expect(vs.Spec.Http[0].Headers.Request.Set).To(HaveLen(3))
					Expect(vs.Spec.Http[0].Headers.Request.Set["Cookie"]).To(Equal("header1=value1"))
					Expect(vs.Spec.Http[0].Headers.Request.Set["header2"]).To(Equal("value2"))
					Expect(vs.Spec.Http[0].Headers.Request.Set["x-forwarded-host"]).To(Equal("example-host.example.com"))
				},
			}, nil, "create"),
	)
})
