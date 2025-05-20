package virtualservice_test

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	processors "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1/virtualservice"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("GetVirtualServiceHttpTimeout", func() {
	It("should return default of 180s when no timeout is set", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{}
		rule := gatewayv2alpha1.Rule{}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(180)))
	})

	It("should return the timeout set in the rule when it is set and APIRule has different value", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(20)),
		}
		rule := gatewayv2alpha1.Rule{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(10)),
		}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(10)))
	})

	It("should return the timeout set in the rule when it is set and APIRule timeout is not", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{}
		rule := gatewayv2alpha1.Rule{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(10)),
		}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(10)))
	})

	It("should return the timeout set in the APIRule it is set and rule timeout is not", func() {
		// given
		apiRuleSpec := gatewayv2alpha1.APIRuleSpec{
			Timeout: ptr.To(gatewayv2alpha1.Timeout(20)),
		}
		rule := gatewayv2alpha1.Rule{}

		// when
		timeout := processors.GetVirtualServiceHttpTimeout(apiRuleSpec, rule)

		// then
		Expect(timeout).To(Equal(uint32(20)))
	})
})
