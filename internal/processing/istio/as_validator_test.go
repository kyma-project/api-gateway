package istio

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessStrategies Istio Validator", func() {
	It("Should succeed with only allow handler", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "allow",
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
					Name: "noop",
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed with only jwt handler", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "jwt",
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail with allow and jwt handlers on same path", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "allow",
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "jwt",
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

	It("Should fail with noop and jwt handlers on same path", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "noop",
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "jwt",
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[1].handler"))
		Expect(problems[0].Message).To(Equal("jwt access strategy is not allowed in combination with other access strategies"))
	})

	It("Should fail with jwt and noop handlers on same path, reverse order", func() {
		//given
		strategies := []*gatewayv1beta1.Authenticator{
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "jwt",
				},
			},
			{
				Handler: &gatewayv1beta1.Handler{
					Name: "noop",
				},
			},
		}
		//when
		problems := (&asValidator{}).Validate("some.attribute", strategies)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.accessStrategies[0].handler"))
		Expect(problems[0].Message).To(Equal("jwt access strategy is not allowed in combination with other access strategies"))
	})
})
