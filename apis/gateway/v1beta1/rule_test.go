package v1beta1_test

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rule", func() {

	Describe("HasRestrictedMethodAccess", func() {

		It("should return false when access strategy allow is defined in first place", func() {
			rule := v1beta1.Rule{
				AccessStrategies: []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyAllow,
						},
					},
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyJwt,
						},
					},
				},
			}

			Expect(rule.HasRestrictedMethodAccess()).To(BeFalse())
		})

		It("should return false when access strategy allow is defined as last in the list", func() {
			rule := v1beta1.Rule{
				AccessStrategies: []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyJwt,
						},
					},
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyAllow,
						},
					},
				},
			}

			Expect(rule.HasRestrictedMethodAccess()).To(BeFalse())
		})

		It("should return true when access strategy does not include allow", func() {
			rule := v1beta1.Rule{
				AccessStrategies: []*v1beta1.Authenticator{
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyJwt,
						},
					},
					{
						Handler: &v1beta1.Handler{
							Name: v1beta1.AccessStrategyAllowMethods,
						},
					},
				},
			}

			Expect(rule.HasRestrictedMethodAccess()).To(BeTrue())
		})

	})
})
