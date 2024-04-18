package v1beta2_test

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rule", func() {

	Describe("ContainsAccessStrategy", func() {

		It("should return false when noAuth does not exist in the object", func() {
			rule := v1beta2.Rule{
				Jwt: &v1beta2.JwtConfig{},
			}

			Expect(rule.ContainsNoAuth()).To(BeFalse())
		})

		It("should return true when noAuth exists in the object", func() {
			value := true

			rule := v1beta2.Rule{
				NoAuth: &value,
			}

			Expect(rule.ContainsNoAuth()).To(BeTrue())
		})

		It("should return true when JWT exists in the object", func() {
			rule := v1beta2.Rule{
				Jwt: &v1beta2.JwtConfig{},
			}

			Expect(rule.ContainsAccessStrategyJwt()).To(BeTrue())
		})

		It("should return false when no access strategy is in the object", func() {
			rule := v1beta2.Rule{}

			Expect(rule.ContainsAccessStrategyJwt()).To(BeFalse())
			Expect(rule.ContainsNoAuth()).To(BeFalse())
		})

	})
})
