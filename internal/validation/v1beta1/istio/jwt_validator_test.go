package istio

import (
	"encoding/json"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	processingtest "github.com/kyma-project/api-gateway/internal/processing/processing_test"
	"github.com/kyma-project/api-gateway/internal/types/ory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("JWT Handler validation", func() {

	It("Should fail with empty config", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: emptyJWTIstioConfig()}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
		Expect(problems[0].Message).To(Equal("supplied config cannot be empty"))
	})

	Context("JWT authentication configuration", func() {
		It("Should fail validation when jwksUri is not a URI", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("no_url", "the_issuer")}

			//when
			problems := (&handlerValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.authentications[0].jwksUri"))
			Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid uri"))
		})

		It("Should fail validation when issuer and jwksUri are empty", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("", "")}

			//when
			problems := (&handlerValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(2))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.authentications[0].issuer"))
			Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid uri"))
			Expect(problems[1].AttributePath).To(Equal("some.attribute.config.authentications[0].jwksUri"))
			Expect(problems[1].Message).To(ContainSubstring("value is empty or not a valid uri"))
		})

		It("Should fail validation when issuer contains ':' but is not a URI", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("http://issuer.test/.well-known/jwks.json", "://example")}

			//when
			problems := (&handlerValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.authentications[0].issuer"))
			Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid uri"))
		})

		It("Should pass validation when issuer contains ':' and is a URI", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("http://issuer.test/.well-known/jwks.json", "https://issuer.example.com")}

			//when
			problems := (&handlerValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should pass validation when issuer is not empty has contains no ':'", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("http://issuer.test/.well-known/jwks.json", "testing@secure.istio.io")}

			//when
			problems := (&handlerValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})

	It("Should succeed for config with plain HTTP JWKSUrls and trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("http://issuer.test/.well-known/jwks.json", "http://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for config with file JWKSUrls and HTTPS trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("file://.well-known/jwks.json", "https://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for config with HTTPS JWKSUrls and trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: jwtIstioConfig("https://issuer.test/.well-known/jwks.json", "https://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for invalid JSON", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: &runtime.RawExtension{Raw: []byte("/abc]")}}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
		Expect(problems[0].Message).To(Equal("Can't read json: invalid character '/' looking for beginning of value"))
	})

	It("Should fail for config with Ory JWT configuration", func() {
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTOryConfig("https://issuer.test/.well-known/jwks.json", "https://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(Not(BeEmpty()))
	})

	Context("for authentications", func() {

		It("Should fail validation when authentications are empty", func() {
			//given
			config := processingtest.GetRawConfig(
				gatewayv1beta1.JwtConfig{
					Authorizations: []*gatewayv1beta1.JwtAuthorization{
						{
							RequiredScopes: []string{"scope-a", "scope-b"},
						},
					},
				})

			handler := &gatewayv1beta1.Handler{
				Name:   "jwt",
				Config: config,
			}

			//when
			problems := (&handlerValidator{}).Validate("", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].Message).To(Equal("Authentications are required when using JWT access handler"))
		})

		It("Should fail validation when authentication has more than one fromHeaders", func() {
			//given
			config := processingtest.GetRawConfig(
				gatewayv1beta1.JwtConfig{
					Authentications: []*gatewayv1beta1.JwtAuthentication{
						{
							Issuer:      "https://issuer.test/",
							JwksUri:     "file://.well-known/jwks.json",
							FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header1"}, {Name: "header2"}},
						},
					},
				})

			handler := &gatewayv1beta1.Handler{
				Name:   "jwt",
				Config: config,
			}

			//when
			problems := (&handlerValidator{}).Validate("", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".config.authentications[0].fromHeaders"))
			Expect(problems[0].Message).To(Equal("multiple fromHeaders are not supported"))
		})

		It("Should fail validation when authentication has more than one fromParams", func() {
			//given
			config := processingtest.GetRawConfig(
				gatewayv1beta1.JwtConfig{
					Authentications: []*gatewayv1beta1.JwtAuthentication{
						{
							Issuer:     "https://issuer.test/",
							JwksUri:    "file://.well-known/jwks.json",
							FromParams: []string{"param1", "param2"},
						},
					},
				})

			handler := &gatewayv1beta1.Handler{
				Name:   "jwt",
				Config: config,
			}

			//when
			problems := (&handlerValidator{}).Validate("", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".config.authentications[0].fromParams"))
			Expect(problems[0].Message).To(Equal("multiple fromParams are not supported"))
		})

		It("Should fail validation when multiple authentications have mixture of fromHeaders and fromParams", func() {
			//given
			config := processingtest.GetRawConfig(
				gatewayv1beta1.JwtConfig{
					Authentications: []*gatewayv1beta1.JwtAuthentication{
						{
							Issuer:     "https://issuer.test/",
							JwksUri:    "file://.well-known/jwks.json",
							FromParams: []string{"param1"},
						},
						{
							Issuer:      "https://issuer.test/",
							JwksUri:     "file://.well-known/jwks.json",
							FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header1"}},
						},
					},
				})

			handler := &gatewayv1beta1.Handler{
				Name:   "jwt",
				Config: config,
			}

			//when
			problems := (&handlerValidator{}).Validate("", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".config.authentications[1].fromHeaders"))
			Expect(problems[0].Message).To(Equal("mixture of multiple fromHeaders and fromParams is not supported"))
		})
	})

	Context("for authorizations", func() {

		It("Should fail validation when authorization has no value", func() {
			//given
			config := processingtest.GetRawConfig(
				gatewayv1beta1.JwtConfig{
					Authentications: []*gatewayv1beta1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
					Authorizations: []*gatewayv1beta1.JwtAuthorization{
						nil,
					},
				})

			handler := &gatewayv1beta1.Handler{
				Name:   "jwt",
				Config: config,
			}

			//when
			problems := (&handlerValidator{}).Validate("", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".config.authorizations[0]"))
			Expect(problems[0].Message).To(Equal("authorization is empty"))
		})

		It("Should successful validate config without authorizations", func() {
			//given
			config := processingtest.GetRawConfig(
				gatewayv1beta1.JwtConfig{
					Authentications: []*gatewayv1beta1.JwtAuthentication{
						{
							Issuer:  "https://issuer.test/",
							JwksUri: "file://.well-known/jwks.json",
						},
					},
				})

			handler := &gatewayv1beta1.Handler{
				Name:   "jwt",
				Config: config,
			}

			//when
			problems := (&handlerValidator{}).Validate("", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		Context("required scopes", func() {
			It("Should fail for config with empty required scopes", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						RequiredScopes: []string{},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("some.attribute", handler)

				//then
				Expect(problems).To(HaveLen(1))
				Expect(problems[0].AttributePath).To(Equal("some.attribute.config.authorizations[0].requiredScopes"))
				Expect(problems[0].Message).To(Equal("value is empty"))
			})

			It("Should fail for config with empty string in required scopes", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						RequiredScopes: []string{"scope-a", ""},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("some.attribute", handler)

				//then
				Expect(problems).To(HaveLen(1))
				Expect(problems[0].AttributePath).To(Equal("some.attribute.config.authorizations[0].requiredScopes"))
				Expect(problems[0].Message).To(Equal("scope value is empty"))
			})

			It("Should succeed for config with two required scopes", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						RequiredScopes: []string{"scope-a", "scope-b"},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("some.attribute", handler)

				//then
				Expect(problems).To(HaveLen(0))
			})

			It("Should successful validate config without a required scope", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						Audiences: []string{"www.example.com"},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("", handler)

				//then
				Expect(problems).To(HaveLen(0))
			})
		})

		Context("audiences", func() {

			It("Should fail validation for config with empty audiences", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						Audiences: []string{},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("", handler)

				//then
				Expect(problems).To(HaveLen(1))
				Expect(problems[0].AttributePath).To(Equal(".config.authorizations[0].audiences"))
				Expect(problems[0].Message).To(Equal("value is empty"))
			})

			It("Should fail validation for config with empty string in audiences", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						Audiences: []string{"www.example.com", ""},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("", handler)

				//then
				Expect(problems).To(HaveLen(1))
				Expect(problems[0].AttributePath).To(Equal(".config.authorizations[0].audiences"))
				Expect(problems[0].Message).To(Equal("audience value is empty"))
			})

			It("Should successful validate config with an audience", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						Audiences: []string{"www.example.com"},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("", handler)

				//then
				Expect(problems).To(HaveLen(0))
			})

			It("Should successful validate config without audiences", func() {
				//given
				authorizations := []*gatewayv1beta1.JwtAuthorization{
					{
						RequiredScopes: []string{"www.example.com"},
					},
				}
				handler := &gatewayv1beta1.Handler{
					Name:   "jwt",
					Config: testURLJWTIstioConfigWithAuthorizations(authorizations),
				}

				//when
				problems := (&handlerValidator{}).Validate("", handler)

				//then
				Expect(problems).To(HaveLen(0))
			})
		})
	})

	Context("for rules validation", func() {

		It("Should fail validation when multiple authentications fromHeaders configuration", func() {
			//given
			ruleFromHeaders := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:      "https://issuer.test/",
										JwksUri:     "file://.well-known/jwks.json",
										FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header1"}},
									},
									{
										Issuer:      "https://issuer.test/",
										JwksUri:     "file://.well-known/jwks.json",
										FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header2"}},
									},
								},
							}),
					}},
				},
			}

			//when
			problems := (&RulesValidator{}).Validate(".spec.rules", []gatewayv1beta1.Rule{ruleFromHeaders})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].accessStrategy[0].config.authentications[1]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("Should fail validation when multiple authentications fromParams configuration", func() {
			//given
			ruleFromHeaders := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
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
							}),
					}},
				},
			}

			//when
			problems := (&RulesValidator{}).Validate(".spec.rules", []gatewayv1beta1.Rule{ruleFromHeaders})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[0].accessStrategy[0].config.authentications[1]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("Should fail when multiple jwt handlers specify different token from types of configurations", func() {
			//given
			ruleFromHeaders := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:      "https://issuer.test/",
										JwksUri:     "file://.well-known/jwks.json",
										FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header1"}},
									},
								},
							}),
					}},
				},
			}

			ruleFromParams := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:     "https://issuer.test/",
										JwksUri:    "file://.well-known/jwks.json",
										FromParams: []string{"param1"},
									},
								},
							}),
					}},
				},
			}

			//when
			problems := (&RulesValidator{}).Validate(".spec.rules", []gatewayv1beta1.Rule{ruleFromHeaders, ruleFromParams})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].accessStrategy[0].config.authentications[0]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("Should fail when multiple jwt handlers specify different token from headers configuration", func() {
			//given
			ruleFromHeaders := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:      "https://issuer.test/",
										JwksUri:     "file://.well-known/jwks.json",
										FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header1"}},
									},
								},
							}),
					}},
				},
			}

			ruleFromParams := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:      "https://issuer.test/",
										JwksUri:     "file://.well-known/jwks.json",
										FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header2"}},
									},
								},
							}),
					}},
				},
			}

			//when
			problems := (&RulesValidator{}).Validate(".spec.rules", []gatewayv1beta1.Rule{ruleFromHeaders, ruleFromParams})

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal(".spec.rules[1].accessStrategy[0].config.authentications[0]"))
			Expect(problems[0].Message).To(Equal("multiple jwt configurations that differ for the same issuer"))
		})

		It("Should succeed when multiple jwt handlers specify same token from headers configuration", func() {
			//given
			ruleFromHeaders := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:      "https://issuer.test/",
										JwksUri:     "file://.well-known/jwks.json",
										FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header1"}},
									},
								},
							}),
					}},
				},
			}

			ruleFromParams := gatewayv1beta1.Rule{
				AccessStrategies: []*gatewayv1beta1.Authenticator{{
					Handler: &gatewayv1beta1.Handler{
						Name: "jwt",
						Config: getRawConfig(
							gatewayv1beta1.JwtConfig{
								Authentications: []*gatewayv1beta1.JwtAuthentication{
									{
										Issuer:      "https://issuer.test/",
										JwksUri:     "file://.well-known/jwks.json",
										FromHeaders: []*gatewayv1beta1.JwtHeader{{Name: "header1"}},
									},
								},
							}),
					}},
				},
			}

			//when
			problems := (&RulesValidator{}).Validate(".spec.rules", []gatewayv1beta1.Rule{ruleFromHeaders, ruleFromParams})

			//then
			Expect(problems).To(HaveLen(0))
		})
	})
})

func emptyJWTIstioConfig() *runtime.RawExtension {
	return processingtest.GetRawConfig(
		&gatewayv1beta1.JwtConfig{})
}

func jwtIstioConfig(jwksUri string, issuer string) *runtime.RawExtension {
	return processingtest.GetRawConfig(
		gatewayv1beta1.JwtConfig{
			Authentications: []*gatewayv1beta1.JwtAuthentication{
				{
					Issuer:  issuer,
					JwksUri: jwksUri,
				},
			},
		})
}

func testURLJWTIstioConfigWithAuthorizations(authorizations []*gatewayv1beta1.JwtAuthorization) *runtime.RawExtension {
	return processingtest.GetRawConfig(
		gatewayv1beta1.JwtConfig{
			Authentications: []*gatewayv1beta1.JwtAuthentication{
				{
					Issuer:  "https://issuer.test/",
					JwksUri: "file://.well-known/jwks.json",
				},
			},
			Authorizations: authorizations,
		})
}

func testURLJWTOryConfig(JWKSUrls string, trustedIssuers string) *runtime.RawExtension {
	return processingtest.GetRawConfig(
		&ory.JWTAccStrConfig{
			JWKSUrls:       []string{JWKSUrls},
			TrustedIssuers: []string{trustedIssuers},
			RequiredScopes: []string{"atgo"},
		})
}

func getRawConfig(config any) *runtime.RawExtension {
	b, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: b,
	}
}
