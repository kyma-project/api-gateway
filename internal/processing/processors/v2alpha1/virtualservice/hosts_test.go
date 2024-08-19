package virtualservice_test

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/kyma-project/api-gateway/internal/builders/builders_test/v2alpha1_test"
	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
			NewAPIRuleBuilder().WithGateway("example/example").WithHost("example.com").Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Hosts).To(ConsistOf("example.com"))
				},
			}, "create"),

		Entry("should set multiple hosts correctly",
			NewAPIRuleBuilder().WithGateway("example/example").WithHosts("example.com", "goat.com").Build(),
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Hosts).To(ConsistOf("example.com", "goat.com"))
				},
			}, "create"),
	)
})
