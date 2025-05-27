package v2alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

var _ = Describe("JWT authorizations validation", func() {

	It("should fail validation when authorizations is defined, but empty", func() {
		//given
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "https://issuer.test/",
						JwksUri: "file://.well-known/jwks.json",
					},
				},
				Authorizations: []*gatewayv2alpha1.JwtAuthorization{},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authorizations"))
		Expect(problems[0].Message).To(Equal("authorizations defined, but no configuration exists"))
	})

	It("should fail validation when authorization is empty", func() {
		//given
		var auth *gatewayv2alpha1.JwtAuthorization
		rule := gatewayv2alpha1.Rule{
			Jwt: &gatewayv2alpha1.JwtConfig{
				Authentications: []*gatewayv2alpha1.JwtAuthentication{
					{
						Issuer:  "https://issuer.test/",
						JwksUri: "file://.well-known/jwks.json",
					},
				},
				Authorizations: []*gatewayv2alpha1.JwtAuthorization{
					auth,
				},
			},
		}

		//when
		problems := validateJwt("rule", &rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("rule.jwt.authorizations[0]"))
		Expect(problems[0].Message).To(Equal("authorization is empty"))
	})

	Context("required scopes", func() {
		It("should fail for config with empty required scopes", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							RequiredScopes: []string{},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("rule.jwt.authorizations[0].requiredScopes"))
			Expect(problems[0].Message).To(Equal("value is empty"))
		})

		It("should fail for config with empty string in required scopes", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							RequiredScopes: []string{"scope-a", ""},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("rule.jwt.authorizations[0].requiredScopes"))
			Expect(problems[0].Message).To(Equal("scope value is empty"))
		})

		It("should succeed for config with two required scopes", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							RequiredScopes: []string{"scope-a", "scope-b"},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("should successful validate config without a required scope", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							Audiences: []string{"www.example.com"},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})

	Context("audiences", func() {

		It("should fail validation for config with empty audiences", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							Audiences: []string{},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("rule.jwt.authorizations[0].audiences"))
			Expect(problems[0].Message).To(Equal("value is empty"))
		})

		It("should fail validation for config with empty string in audiences", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							Audiences: []string{"www.example.com", ""},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("rule.jwt.authorizations[0].audiences"))
			Expect(problems[0].Message).To(Equal("audience value is empty"))
		})

		It("should successful validate config with an audience", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							Audiences: []string{"www.example.com"},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("should successful validate config without audiences", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
					Authentications: []*gatewayv2alpha1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv2alpha1.JwtAuthorization{
						{
							RequiredScopes: []string{"www.example.com"},
						},
					},
				},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})
})
