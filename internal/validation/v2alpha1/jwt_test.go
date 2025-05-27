package v2alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

var _ = Describe("validateJwt", func() {

	It("should fail with empty JWT config", func() {
		// given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{},
		}

		// when
		problems := validateJwt("rule", &rule)

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications"))
		Expect(problems[0].Message).To(Equal("A JWT config must have at least one authentication"))
	})

	It("should fail when Authorizations are configured, but empty Authentications", func() {
		// given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authorizations: []*gatewayv2alpha1.JwtAuthorization{
					{
						RequiredScopes: []string{"scope-a", "scope-b"},
					},
				},
			},
		}

		// when
		problems := validateJwt("rule", &rule)

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications"))
		Expect(problems[0].Message).To(Equal("A JWT config must have at least one authentication"))
	})
})

var _ = Describe("validateJwtAuthenticationEquality", func() {
	It("should fail validation when multiple authentications fromHeaders configuration", func() {
		// given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header1"}},
					},
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header2"}},
					},
				},
			},
		}

		// when
		problems := validateJwtAuthenticationEquality([]gatewayv2alpha1.Rule{rule})

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].jwt.authentications[1]"))
		Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
	})

	It("should fail validation when multiple authentications fromParams configuration", func() {
		// given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:     "https://issuer.test/",
						JwksUri:    "file://.well-known/jwks.json",
						FromParams: []string{"param1"},
					},
					{
						Issuer:     "https://issuer.test/",
						JwksUri:    "file://.well-known/jwks.json",
						FromParams: []string{"param2"},
					},
				},
			},
		}

		// when
		problems := validateJwtAuthenticationEquality([]gatewayv2alpha1.Rule{rule})

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].jwt.authentications[1]"))
		Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
	})

	It("should fail when multiple jwt handlers specify different token from types of configurations", func() {
		// given
		ruleFromHeaders := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header1"}},
					},
				},
			},
		}

		ruleFromParams := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:     "https://issuer.test/",
						JwksUri:    "file://.well-known/jwks.json",
						FromParams: []string{"param1"},
					},
				},
			},
		}

		// when
		problems := validateJwtAuthenticationEquality([]gatewayv2alpha1.Rule{ruleFromHeaders, ruleFromParams})

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].jwt.authentications[0]"))
		Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
	})

	It("should fail when multiple jwt handlers specify different token from headers configuration", func() {
		// given
		ruleFromHeaders := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header1"}},
					},
				},
			},
		}

		ruleFromHeadersDifferent := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header2"}},
					},
				},
			},
		}

		// when
		problems := validateJwtAuthenticationEquality([]gatewayv2alpha1.Rule{ruleFromHeaders, ruleFromHeadersDifferent})

		// then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].jwt.authentications[0]"))
		Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
	})

	It("should succeed when multiple jwt handlers specify same token from headers configuration", func() {
		// given
		ruleFromHeaders := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header1"}},
					},
				},
			},
		}

		ruleFromHeadersEqual := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header1"}},
					},
				},
			},
		}

		// when
		problems := validateJwtAuthenticationEquality([]gatewayv2alpha1.Rule{ruleFromHeaders, ruleFromHeadersEqual})

		// then
		Expect(problems).To(HaveLen(0))
	})
})
