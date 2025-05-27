package ory

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
)

var _ = Describe("AccessStrategies Istio Validator", func() {

	DescribeTable("Should succeed with only one exclusive handler", func(handler string) {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: handler,
				},
			},
		}
		//when
		problems := (&accessStrategyValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	},
		Entry(nil, gatewayv1beta1.AccessStrategyNoAuth),
		Entry(nil, gatewayv1beta1.AccessStrategyAllow),
	)

	DescribeTable("Should return failures with not exclusive handler oauth2_introspection and with exclusive handler", func(handler string, expectedMessage string) {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyOauth2Introspection,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: handler,
				},
			},
		}
		//when
		problems := (&accessStrategyValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[0].Message).To(Equal(expectedMessage))
	},
		Entry(nil, gatewayv1beta1.AccessStrategyAllow, "allow access strategy is not allowed in combination with other access strategies"),
		Entry(nil, gatewayv1beta1.AccessStrategyNoAuth, "no_auth access strategy is not allowed in combination with other access strategies"),
		Entry(nil, gatewayv1beta1.AccessStrategyNoop, "noop access strategy is not allowed in combination with other access strategies"),
	)

	It("Should return multiple failures when there are multiple exclusive handlers", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllow,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoAuth,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoop,
				},
			},
		}
		//when
		problems := (&accessStrategyValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(3))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[0].handler"))
		Expect(problems[0].Message).To(Equal("allow access strategy is not allowed in combination with other access strategies"))
		Expect(problems[1].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[1].Message).To(Equal("no_auth access strategy is not allowed in combination with other access strategies"))
		Expect(problems[2].AttributePath).To(Equal("some.attribute.accessStrategies[2].handler"))
		Expect(problems[2].Message).To(Equal("noop access strategy is not allowed in combination with other access strategies"))
	})

	It("Should succeed with multiple non-exclusive handler", func() {
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
		}
		//when
		problems := (&accessStrategyValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	})
})
