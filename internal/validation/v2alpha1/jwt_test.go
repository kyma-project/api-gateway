package v2alpha1

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JWT validation", func() {

	Context("validateJwt", func() {

		It("should fail with empty JWT config", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{},
			}

			//when
			problems := validateJwt("rule", &rule)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications"))
			Expect(problems[0].Message).To(Equal("A JWT config must have at least one authentication"))
		})

		It("should fail when Authorizations are configured, but empty Authentications", func() {
			//given
			rule := gatewayv2alpha1.Rule{
				Jwt: &gatewayv2alpha1.JwtConfig{
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
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("rule.jwt.authentications"))
			Expect(problems[0].Message).To(Equal("A JWT config must have at least one authentication"))
		})

		Context("for authentications", func() {

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

		Context("for authorizations", func() {

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
	})

	Context("validateJwtAuthenticationEquality", func() {

		It("should fail validation when multiple authentications fromHeaders configuration", func() {
			//given
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

			//when
			problems := validateJwtAuthenticationEquality(".spec.rules", []gatewayv2alpha1.Rule{rule})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].jwt.authentications[1]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("should fail validation when multiple authentications fromParams configuration", func() {
			//given
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

			//when
			problems := validateJwtAuthenticationEquality(".spec.rules", []gatewayv2alpha1.Rule{rule})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].jwt.authentications[1]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("should fail when multiple jwt handlers specify different token from types of configurations", func() {
			//given
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

			//when
			problems := validateJwtAuthenticationEquality(".spec.rules", []gatewayv2alpha1.Rule{ruleFromHeaders, ruleFromParams})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].jwt.authentications[0]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("should fail when multiple jwt handlers specify different token from headers configuration", func() {
			//given
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

			//when
			problems := validateJwtAuthenticationEquality(".spec.rules", []gatewayv2alpha1.Rule{ruleFromHeaders, ruleFromHeadersDifferent})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].jwt.authentications[0]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("should succeed when multiple jwt handlers specify same token from headers configuration", func() {
			//given
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

			//when
			problems := validateJwtAuthenticationEquality(".spec.rules", []gatewayv2alpha1.Rule{ruleFromHeaders, ruleFromHeadersEqual})

			//then
			Expect(problems).To(HaveLen(0))
		})
	})

})
