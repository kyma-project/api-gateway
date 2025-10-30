package v2alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"

	. "github.com/kyma-project/api-gateway/internal/processing/processing_test"
)

var _ = Describe("StatusBase", func() {
	It("should create status from given status code", func() {
		// given
		fakeClient := GetFakeClient()
		r := v2alpha1.NewReconciliation(nil, nil, nil, nil, processing.ReconciliationConfig{}, nil, false, fakeClient)

		// when
		s, ok := r.GetStatusBase(string(gatewayv2alpha1.Error)).(status.ReconciliationV2alpha1Status)
		// then
		Expect(ok).Should(BeTrue())
		Expect(s.ApiRuleStatus.State).Should(Equal(gatewayv2alpha1.Error))
		Expect(s.ApiRuleStatus.Description).Should(BeEmpty())
	})
})
