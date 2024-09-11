package v2alpha1_test

import (
	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	apirulev2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	v2alpha1 "github.com/kyma-project/api-gateway/internal/builders/builders_test/v2alpha1_test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const annotationKey = "gateway.kyma-project.io/v2alpha1-rules"

var dummyExtAuthRule = v2alpha1.NewRuleBuilder().
	WithPath("/test").
	WithMethods("GET").
	WithExtAuth(v2alpha1.NewExtAuthBuilder().
		WithAuthorizers("test-authorizer").
		WithRestriction(&apirulev2alpha1.JwtConfig{
			Authentications: []*apirulev2alpha1.JwtAuthentication{
				{
					Issuer:  "test-issuer",
					JwksUri: "test-jwks-uri",
					FromHeaders: []*apirulev2alpha1.JwtHeader{
						{
							Name:   "test-header",
							Prefix: "test-prefix",
						},
					},
					FromParams: []string{"test-param"},
				},
			},
			Authorizations: nil,
		}).
		Build()).
	Build()

var _ = Describe("ExtAuthStorage", func() {
	It("Should store extAuth in v1beta1 through annotation and a rule with only handler name set", func() {
		// given
		v2alpha1APIRule := v2alpha1.NewAPIRuleBuilderWithDummyData().WithRules(dummyExtAuthRule).Build()

		// when
		var betaConverted apirulev1beta1.APIRule
		err := v2alpha1APIRule.ConvertTo(&betaConverted)
		Expect(err).ToNot(HaveOccurred())

		//then
		annotations := betaConverted.GetAnnotations()
		Expect(annotations).To(HaveKey(annotationKey))
		Expect(betaConverted.Spec.Rules).To(HaveLen(1))
		Expect(betaConverted.Spec.Rules[0].Path).To(Equal("/test"))
		Expect(betaConverted.Spec.Rules[0].Methods).To(BeEquivalentTo([]apirulev1beta1.HttpMethod{"GET"}))
		Expect(betaConverted.Spec.Rules[0].AccessStrategies).To(HaveLen(1))
		Expect(betaConverted.Spec.Rules[0].AccessStrategies[0].Handler.Name).To(BeEquivalentTo("ext-auth"))
		Expect(betaConverted.Spec.Rules[0].AccessStrategies[0].Config).To(BeNil())
	})
})

var _ = Describe("ExtAuthConversion", func() {

	DescribeTable("Should convert back and forth correctly with ExtAuth set", func(expectedRules []*apirulev2alpha1.Rule) {
		// given
		v2alpha1APIRule := v2alpha1.NewAPIRuleBuilderWithDummyData().WithRules(expectedRules...).Build()
		var betaConverted apirulev1beta1.APIRule
		err := v2alpha1APIRule.ConvertTo(&betaConverted)
		Expect(err).ToNot(HaveOccurred())

		// when
		var v2alpha1ConvertedRule apirulev2alpha1.APIRule
		err = v2alpha1ConvertedRule.ConvertFrom(&betaConverted)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(v2alpha1ConvertedRule.Spec.Rules).To(HaveLen(len(expectedRules)))
		for i, rule := range v2alpha1ConvertedRule.Spec.Rules {
			Expect(rule.Path).To(Equal(expectedRules[i].Path))
			Expect(rule.Methods).To(BeEquivalentTo(expectedRules[i].Methods))
			Expect(rule.Service).To(BeEquivalentTo(expectedRules[i].Service))
			Expect(rule.NoAuth).To(Equal(expectedRules[i].NoAuth))
			Expect(rule.Jwt != nil).To(Equal(expectedRules[i].Jwt != nil))
			if rule.Jwt != nil {
				Expect(rule.Jwt.Authorizations).To(BeEquivalentTo(expectedRules[i].Jwt.Authorizations))
				Expect(rule.Jwt.Authentications).To(BeEquivalentTo(expectedRules[i].Jwt.Authentications))
			}
			Expect(rule.ExtAuth != nil).To(Equal(expectedRules[i].ExtAuth != nil))
			if rule.ExtAuth != nil {
				Expect(rule.ExtAuth.ExternalAuthorizers).To(BeEquivalentTo(expectedRules[i].ExtAuth.ExternalAuthorizers))
				Expect(rule.ExtAuth.Restrictions != nil).To(Equal(expectedRules[i].ExtAuth.Restrictions != nil))
				if rule.ExtAuth.Restrictions != nil {
					Expect(rule.ExtAuth.Restrictions.Authentications).To(BeEquivalentTo(expectedRules[i].ExtAuth.Restrictions.Authentications))
					Expect(rule.ExtAuth.Restrictions.Authorizations).To(BeEquivalentTo(expectedRules[i].ExtAuth.Restrictions.Authorizations))
				}
			}
		}
	},
		Entry("Should convert APIRule with no ExtAuth", []*apirulev2alpha1.Rule{}),
		Entry("Should convert APIRule with only ExtAuth", []*apirulev2alpha1.Rule{dummyExtAuthRule}),
		Entry("Should preserve order of rules when ExtAuth is in the middle", []*apirulev2alpha1.Rule{
			v2alpha1.NewRuleBuilder().
				WithPath("/first").
				NoAuth().
				Build(),
			dummyExtAuthRule,
			v2alpha1.NewRuleBuilder().
				WithPath("/third").
				NoAuth().
				Build(),
		}),
		Entry("Should preserve order of rules when ExtAuth is at the end", []*apirulev2alpha1.Rule{
			v2alpha1.NewRuleBuilder().
				WithPath("/first").
				NoAuth().
				Build(),
			dummyExtAuthRule,
		}),
		Entry("Should preserve order of rules when ExtAuth is at the beginning", []*apirulev2alpha1.Rule{
			dummyExtAuthRule,
			v2alpha1.NewRuleBuilder().
				WithPath("/second").
				NoAuth().
				Build(),
		}),
	)
})
