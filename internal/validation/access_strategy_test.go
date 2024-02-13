package validation_test

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Access Strategies Validation", func() {

	Describe("CheckForExclusiveAccessStrategy", func() {
		It("Should have no validation failure when there is only one access strategy that is exclusive", func() {
			//given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyNoAuth,
					},
				},
			}

			//when
			problems := validation.CheckForExclusiveAccessStrategy(strategies, gatewayv1beta1.AccessStrategyNoAuth, "some.attribute")

			//then
			Expect(problems).To(HaveLen(0))

		})

		It("Should have no validation failure when there is no access strategy", func() {
			//given
			var strategies []*gatewayv1beta1.Authenticator

			//when
			failure := validation.CheckForExclusiveAccessStrategy(strategies, gatewayv1beta1.AccessStrategyNoAuth, "some.attribute")

			//then
			Expect(failure).To(HaveLen(0))
		})

		It("Should have no validation failure when there only not exclusive access strategies ", func() {
			//given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyJwt,
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyOauth2Introspection,
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyOauth2ClientCredentials,
					},
				},
			}

			//when
			failure := validation.CheckForExclusiveAccessStrategy(strategies, gatewayv1beta1.AccessStrategyNoAuth, "some.attribute")

			//then
			Expect(failure).To(HaveLen(0))
		})

		It("Should validation failure when there is an exclusive access strategy", func() {
			//given
			strategies := []*gatewayv1beta1.Authenticator{
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyJwt,
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyOauth2Introspection,
					},
				},
				{
					Handler: &gatewayv1beta1.Handler{
						Name: gatewayv1beta1.AccessStrategyNoAuth,
					},
				},
			}

			//when
			failure := validation.CheckForExclusiveAccessStrategy(strategies, gatewayv1beta1.AccessStrategyNoAuth, "some.attribute")

			//then
			Expect(failure).To(HaveLen(1))
			Expect(failure[0].AttributePath).To(Equal("some.attribute.accessStrategies[2].handler"))
			Expect(failure[0].Message).To(Equal("no_auth access strategy is not allowed in combination with other access strategies"))
		})

	})
})
