package istio

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessStrategies Istio Validator", func() {

	DescribeTable("Should succeed with only handler", func(handler string) {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: handler,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	},
		Entry(nil, gatewayv1beta1.AccessStrategyAllowMethods),
		Entry(nil, gatewayv1beta1.AccessStrategyAllow),
		Entry(nil, gatewayv1beta1.AccessStrategyNoop),
		Entry(nil, gatewayv1beta1.AccessStrategyJwt),
	)

	DescribeTable("Should fail with handlers on same path: noop and", func(handler string, expectedMessage string) {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoop,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: handler,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[0].Message).To(Equal(expectedMessage))
	},
		Entry(nil, gatewayv1beta1.AccessStrategyAllow, "allow access strategy is not allowed in combination with other access strategies"),
		Entry(nil, gatewayv1beta1.AccessStrategyAllowMethods, "allowMethods access strategy is not allowed in combination with other access strategies"),
		Entry(nil, gatewayv1beta1.AccessStrategyJwt, "jwt access strategy is not allowed in combination with other access strategies"),
	)

	It("Should fail with allow and jwt handlers on same path", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllow,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyJwt,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[0].handler"))
		Expect(problems[0].Message).To(Equal("allow access strategy is not allowed in combination with other access strategies"))
		Expect(problems[1].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[1].Message).To(Equal("jwt access strategy is not allowed in combination with other access strategies"))
	})

	It("Should fail with allowMethods and jwt handlers on same path", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllowMethods,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyJwt,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[0].handler"))
		Expect(problems[0].Message).To(Equal("allowMethods access strategy is not allowed in combination with other access strategies"))
		Expect(problems[1].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[1].Message).To(Equal("jwt access strategy is not allowed in combination with other access strategies"))
	})

	It("Should fail with allowMethods and allow handlers on same path", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllow,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllowMethods,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[0].handler"))
		Expect(problems[0].Message).To(Equal("allow access strategy is not allowed in combination with other access strategies"))
		Expect(problems[1].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[1].Message).To(Equal("allowMethods access strategy is not allowed in combination with other access strategies"))
	})
})
