package v2alpha1_test

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/builders"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"

	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

type mockVirtualServiceCreator struct{}

func (r mockVirtualServiceCreator) Create(_ *gatewayv2alpha1.APIRule) (*networkingv1.VirtualService, error) {
	return builders.VirtualService().Get(), nil
}
