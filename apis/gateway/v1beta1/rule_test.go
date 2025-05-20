package v1beta1_test

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rule", func() {

	Describe("ContainsAccessStrategy", func() {

		It("should return false when access strategy does not exist in the array", func() {
			rule := v1beta1.Rule{
				AccessStrategies: []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyNoAuth,
						},
					},
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyJwt,
						},
					},
				},
			}

			Expect(rule.ContainsAccessStrategy(v1beta1.AccessStrategyOauth2Introspection)).To(BeFalse())
		})

		It("should return true when access strategy exists in the array", func() {
			rule := v1beta1.Rule{
				AccessStrategies: []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyNoAuth,
						},
					},
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyJwt,
						},
					},
				},
			}

			Expect(rule.ContainsAccessStrategy(v1beta1.AccessStrategyNoAuth)).To(BeTrue())
		})

		It("should return false when no access strategy is in the arrray", func() {
			rule := v1beta1.Rule{
				AccessStrategies: []*v1beta1.Authenticator{},
			}

			Expect(rule.ContainsAccessStrategy(v1beta1.AccessStrategyOauth2Introspection)).To(BeFalse())
		})

	})
})
