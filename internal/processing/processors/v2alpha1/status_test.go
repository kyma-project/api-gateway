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
		// VirtualService and AuthorizationPolicy are always managed; RA is not,
		// since the nil APIRule has no JWT/ext-auth rule.
		Expect(s.ManagedResources[status.OnVirtualService]).Should(BeTrue())
		Expect(s.ManagedResources[status.OnAuthorizationPolicy]).Should(BeTrue())
		Expect(s.ManagedResources[status.OnRequestAuthentication]).Should(BeFalse())
	})

	It("should mark RequestAuthentication managed when a rule requests JWT", func() {
		// given
		fakeClient := GetFakeClient()
		apiRule := &gatewayv2alpha1.APIRule{
			Spec: gatewayv2alpha1.APIRuleSpec{
				Rules: []gatewayv2alpha1.Rule{
					{Jwt: &gatewayv2alpha1.JwtConfig{}},
				},
			},
		}
		r := v2alpha1.NewReconciliation(apiRule, nil, nil, nil, processing.ReconciliationConfig{}, nil, false, fakeClient)

		// when
		s, ok := r.GetStatusBase(string(gatewayv2alpha1.Ready)).(status.ReconciliationV2alpha1Status)
		// then
		Expect(ok).Should(BeTrue())
		Expect(s.ManagedResources[status.OnRequestAuthentication]).Should(BeTrue())
	})
})
