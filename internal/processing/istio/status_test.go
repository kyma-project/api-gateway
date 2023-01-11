package istio

import (
	"testing"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIstioStatusBase(t *testing.T) {
	It("should create status base with AccessRule set to nil", func() {
		// when
		status := IstioStatusBase(gatewayv1beta1.StatusError)

		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusError))
		Expect(status.ApiRuleStatus.Description).To(Equal("error during processor execution"))
		Expect(status.AccessRuleStatus.Code).To(BeNil())
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.RequestAuthenticationStatus).To(Equal(gatewayv1beta1.StatusSkipped))
	})

}
