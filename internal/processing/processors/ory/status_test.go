package ory_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	processingt "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/ory"
	status "github.com/kyma-project/api-gateway/internal/processing/status"
)

var _ = Describe("OryStatusBase", func() {
	It("should create status base with AP and RA set to nil", func() {
		fakeClient := processingt.GetFakeClient()
		r := ory.NewOryReconciliation(nil, processing.ReconciliationConfig{}, nil, fakeClient)
		status := r.GetStatusBase(string(gatewayv1beta1.StatusSkipped)).(status.ReconciliationV1beta1Status)

		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AccessRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus).To(BeNil())
		Expect(status.RequestAuthenticationStatus).To(BeNil())
	})
})
