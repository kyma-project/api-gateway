package ory

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OryStatusBase", func() {
	It("should create status base with AP and RA set to nil", func() {
		// when
		status := StatusBase(gatewayv1beta1.StatusSkipped)

		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())
	})
})
