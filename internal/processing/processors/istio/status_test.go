package istio_test

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	processingt "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	status "github.com/kyma-project/api-gateway/internal/processing/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IstioStatusBase", func() {
	It("should create status base with AccessRule set to nil", func() {
		fakeClient := processingt.GetFakeClient()
		r := istio.NewIstioReconciliation(nil, processing.ReconciliationConfig{}, nil, fakeClient)
		status := r.GetStatusBase(string(gatewayv1beta1.StatusSkipped)).(status.ReconciliationV1beta1Status)

		Expect(status.ApiRuleStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.VirtualServiceStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AuthorizationPolicyStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.RequestAuthenticationStatus.Code).To(Equal(gatewayv1beta1.StatusSkipped))
		Expect(status.AccessRuleStatus).To(BeNil())
	})
})
