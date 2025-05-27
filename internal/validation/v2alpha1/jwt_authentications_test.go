package v2alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

var _ = Describe("JWT authentications validation", func() {

	It("should fail validation when jwksUri is not a URI", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "the_issuer/",
						JwksUri: "no_url",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications[0].jwksUri"))
		Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid uri"))
	})

	It("should fail validation when issuer and jwksUri are empty", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "",
						JwksUri: "",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications[0].issuer"))
		Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid uri"))
		Expect(problems[1].AttributePath).To(Equal("rule.jwt.authentications[0].jwksUri"))
		Expect(problems[1].Message).To(ContainSubstring("value is empty or not a valid uri"))
	})

	It("should fail validation when issuer contains ':' but is not a URI", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "://example",
						JwksUri: "http://issuer.test/.well-known/jwks.json",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications[0].issuer"))
		Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid uri"))
	})

	It("should pass validation when issuer contains ':' and is a URI", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "https://issuer.example.com",
						JwksUri: "http://issuer.test/.well-known/jwks.json",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("should pass validation when issuer is not empty has contains no ':'", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "testing@secure.istio.io",
						JwksUri: "http://issuer.test/.well-known/jwks.json",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("should succeed for config with plain HTTP JWKSUrls and trustedIssuers", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "http://issuer.test",
						JwksUri: "http://issuer.test/.well-known/jwks.json",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("should succeed for config with file JWKSUrls and HTTPS trustedIssuers", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "https://issuer.test/",
						JwksUri: "file://.well-known/jwks.json",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("should succeed for config with HTTPS JWKSUrls and trustedIssuers", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "https://issuer.test//",
						JwksUri: "https://issuer.test/.well-known/jwks.json",
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail validation when authentication has more than one fromHeaders", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:      "https://issuer.test//",
						JwksUri:     "https://issuer.test/.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header1"}, {Name: "header2"}},
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications[0].fromHeaders"))
		Expect(problems[0].Message).To(Equal("multiple fromHeaders are not supported"))
	})

	It("should fail validation when authentication has more than one fromParams", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:     "https://issuer.test//",
						JwksUri:    "https://issuer.test/.well-known/jwks.json",
						FromParams: []string{"param1", "param2"},
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications[0].fromParams"))
		Expect(problems[0].Message).To(Equal("multiple fromParams are not supported"))
	})

	It("should fail validation when multiple authentications have mixture of fromHeaders and fromParams", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:     "https://issuer.test//",
						JwksUri:    "file://.well-known/jwks.json",
						FromParams: []string{"param1"},
					},
					{
						Issuer:      "https://issuer.test/",
						JwksUri:     "file://.well-known/jwks.json",
						FromHeaders: []*gatewayv2alpha1.JwtHeader{{Name: "header1"}},
					},
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications[1].fromHeaders"))
		Expect(problems[0].Message).To(Equal("mixture of multiple fromHeaders and fromParams is not supported"))
	})
})
