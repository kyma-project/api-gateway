package istio

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IstioStatusBase", func() {
	It("should create status base with AccessRule set to nil", func() {
		// when
		status := IstioStatusBase(gatewayv1beta1.StatusSkipped)

		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AccessRuleStatus).To(BeNil())
	})
})
