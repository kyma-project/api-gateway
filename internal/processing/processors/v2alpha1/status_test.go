package v2alpha1_test

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatusBase", func() {
	It("should create status from given status code", func() {
		r := v2alpha1.NewReconciliation(nil, nil, nil, nil, processing.ReconciliationConfig{}, nil, false)

		// when
		s, ok := r.GetStatusBase(string(gatewayv2alpha1.Error)).(status.ReconciliationV2alpha1Status)
		Expect(ok).Should(BeTrue())
		Expect(s.ApiRuleStatus.State).Should(Equal(gatewayv2alpha1.Error))
		Expect(s.ApiRuleStatus.Description).Should(BeEmpty())
	})
})
