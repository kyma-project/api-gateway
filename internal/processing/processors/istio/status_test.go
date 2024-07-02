package istio

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	v1beta1Status "github.com/kyma-project/api-gateway/internal/processing/status/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IstioStatusBase", func() {
	It("should create status base with AccessRule set to nil", func() {
		// when
		status := StatusBase(string(gatewayv1beta1.StatusSkipped)).(v1beta1Status.ReconciliationV1beta1Status)

		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AccessRuleStatus).To(BeNil())
	})
})
