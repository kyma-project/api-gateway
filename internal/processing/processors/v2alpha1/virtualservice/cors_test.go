package virtualservice_test

import (
	istioapiv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kyma-project/api-gateway/internal/builders/builders_test/v2alpha1_test"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
)

var _ = Describe("CORS", func() {
	var client client.Client
	var processor processors.VirtualServiceProcessor
	BeforeEach(func() {
		client = GetFakeClient()
	})

	DescribeTable("CORS",
		func(apiRule *gatewayv2alpha1.APIRule, verifiers []verifier, expectedError error, expectedActions ...string) {
			processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule, nil, client)
			checkVirtualServices(client, processor, verifiers, expectedError, expectedActions...)
		},

		Entry("should set default empty values in VirtualService CORSPolicy when no CORS configuration is set in APIRule",
			NewAPIRuleBuilderWithDummyDataWithNoAuthRule().Build(),
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
			}, nil, "create"),

		Entry("should apply all CORSPolicy headers correctly",
			NewAPIRuleBuilderWithDummyDataWithNoAuthRule().WithCORSPolicy(
				NewCorsPolicyBuilder().
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
			}}, nil, "create"),
	)
})
