package ory

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessStrategies Ory Validator", func() {
	It("Should succeed with only allow handler", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllow,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed with only allow_methods handler", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllowMethods,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed with only noop handler", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoop,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail with allow and noop handler", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllow,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoop,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[0].handler"))
		Expect(problems[0].Message).To(Equal("allow access strategy is not allowed in combination with other access strategies"))
	})

	It("Should fail with allow_methods and noop handler", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllowMethods,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoop,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[0].handler"))
		Expect(problems[0].Message).To(Equal("allow_methods access strategy is not allowed in combination with other access strategies"))
	})

	It("Should fail with allow and noop handler, reverse order", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyNoop,
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: gatewayv1beta1.AccessStrategyAllow,
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[0].Message).To(Equal("allow access strategy is not allowed in combination with other access strategies"))
	})
})
