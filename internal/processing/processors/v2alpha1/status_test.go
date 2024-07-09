package v2alpha1_test

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	status "github.com/kyma-project/api-gateway/internal/processing/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatusBase", func() {
	It("should create status from given status code", func() {
		r := v2alpha1.NewReconciliation(nil, nil, processing.ReconciliationConfig{}, nil)

		// when
		status := r.GetStatusBase(string(gatewayv1beta1.StatusSkipped)).(status.ReconciliationV1beta1Status)

		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
	})
})
