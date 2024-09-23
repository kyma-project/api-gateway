package virtualservice_test

import (
	"errors"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
	apinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		func(apiRule *gatewayv2alpha1.APIRule, gateway *networkingv1beta1.Gateway, verifiers []verifier, expectedError error, expectedActions ...string) {
			processor = processors.NewVirtualServiceProcessor(GetTestConfig(), apiRule, gateway)
			checkVirtualServices(client, processor, verifiers, expectedError, expectedActions...)
		},

		Entry("should set the host correctly",
			NewAPIRuleBuilder().WithGateway("gateway-ns/gateway-name").WithHost("example.com").Build(),
			nil,
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Hosts).To(ConsistOf("example.com"))
				},
			}, nil, "create"),

		Entry("should set multiple hosts correctly",
			NewAPIRuleBuilder().WithGateway("gateway-ns/gateway-name").WithHosts("example.com", "goat.com").Build(),
			nil,
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Hosts).To(ConsistOf("example.com", "goat.com"))
				},
			}, nil, "create"),

		Entry("should set the host correctly when short host is used",
			NewAPIRuleBuilder().WithGateway("gateway-ns/gateway-name").WithHost("example").Build(),
			&networkingv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-ns"},
				Spec: apinetworkingv1beta1.Gateway{
					Servers: []*apinetworkingv1beta1.Server{
						{
							Hosts: []string{
								"*.domain.name",
							},
						},
					},
				},
			},
			[]verifier{
				func(vs *networkingv1beta1.VirtualService) {
					Expect(vs.Spec.Hosts).To(ConsistOf("example.domain.name"))
				},
			}, nil, "create"),

		Entry("should return error when short host is used but no gateway available",
			NewAPIRuleBuilder().WithGateway("gateway-ns/gateway-name").WithHost("example").Build(),
			nil,
			[]verifier{}, errors.New("gateway or host definition is missing")),

		Entry("should return error when short host is used but gateway do not have servers defined",
			NewAPIRuleBuilder().WithGateway("gateway-ns/gateway-name").WithHost("example").Build(),
			&networkingv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-ns"},
				Spec: apinetworkingv1beta1.Gateway{
					Servers: []*apinetworkingv1beta1.Server{},
				},
			},
			[]verifier{}, errors.New("gateway or host definition is missing")),

		Entry("should return error when short host is used but gateway do not have hosts defined",
			NewAPIRuleBuilder().WithGateway("gateway-ns/gateway-name").WithHost("example").Build(),
			&networkingv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{Name: "gateway-name", Namespace: "gateway-ns"},
				Spec: apinetworkingv1beta1.Gateway{
					Servers: []*apinetworkingv1beta1.Server{
						{
							Hosts: []string{},
						},
					},
				},
			},
			[]verifier{}, errors.New("gateway or host definition is missing")),
	)
})
