package processors_test

import (
	"context"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	. "github.com/kyma-incubator/api-gateway/internal/processing/internal/test"
	"github.com/kyma-incubator/api-gateway/internal/processing/processors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("Request Authentication Processor", func() {
	It("should create RA when no exists", func() {
		// given
		apiRule := &gatewayv1beta1.APIRule{}

		processor := processors.RequestAuthenticationProcessor{
			Creator: mockRequestAuthenticationCreator{},
		}

		// when
		result, err := processor.EvaluateReconciliation(context.TODO(), GetEmptyFakeClient(), apiRule)

		// then
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("create"))
	})
})

type mockRequestAuthenticationCreator struct {
}

func (r mockRequestAuthenticationCreator) Create(_ *gatewayv1beta1.APIRule) map[string]*securityv1beta1.RequestAuthentication {
	return map[string]*securityv1beta1.RequestAuthentication{
		"<http|https>://myService.myDomain.com<path>": builders.RequestAuthenticationBuilder().Get(),
	}
}
