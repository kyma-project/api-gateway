package istio

import (
	"encoding/json"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
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

	It("Should fail for config with invalid trustedIssuers and JWKSUrls", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: simpleJWTIstioConfig("a t g o")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.config.authentications[0].issuer"))
		Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid url"))
		Expect(problems[1].AttributePath).To(Equal("some.attribute.config.authentications[0].jwksUri"))
		Expect(problems[1].Message).To(ContainSubstring("value is empty or not a valid url"))
	})

	It("Should fail for config with plain HTTP JWKSUrls and trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTIstioConfig("http://issuer.test/.well-known/jwks.json", "http://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.config.authentications[0].issuer"))
		Expect(problems[0].Message).To(ContainSubstring("value is not a secured url"))
		Expect(problems[1].AttributePath).To(Equal("some.attribute.config.authentications[0].jwksUri"))
		Expect(problems[1].Message).To(ContainSubstring("value is not a secured url"))
	})

	It("Should succeed for config with file JWKSUrls and HTTPS trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTIstioConfig("file://.well-known/jwks.json", "https://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for config with HTTPS JWKSUrls and trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTIstioConfig("https://issuer.test/.well-known/jwks.json", "https://issuer.test/")}

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

	Context("for authorizations", func() {

		It("Should successful validate config without authorizations", func() {
			//given
			config := getRawConfig(
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
			It("Should have failed validations for config with empty audiences", func() {
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

			It("Should have failed validations for config with empty string in audiences", func() {
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
})

func emptyJWTIstioConfig() *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1beta1.JwtConfig{})
}

func simpleJWTIstioConfig(trustedIssuers ...string) *runtime.RawExtension {
	var issuers []*gatewayv1beta1.JwtAuthentication
	for _, issuer := range trustedIssuers {
		issuers = append(issuers, &gatewayv1beta1.JwtAuthentication{
			Issuer:  issuer,
			JwksUri: issuer,
		})
	}
	jwtConfig := gatewayv1beta1.JwtConfig{Authentications: issuers}
	return getRawConfig(jwtConfig)
}

func testURLJWTIstioConfig(JWKSUrl string, trustedIssuer string) *runtime.RawExtension {
	return getRawConfig(
		gatewayv1beta1.JwtConfig{
			Authentications: []*gatewayv1beta1.JwtAuthentication{
				{
					Issuer:  trustedIssuer,
					JwksUri: JWKSUrl,
				},
			},
		})
}

func testURLJWTIstioConfigWithAuthorizations(authorizations []*gatewayv1beta1.JwtAuthorization) *runtime.RawExtension {
	return getRawConfig(
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
	return getRawConfig(
		&ory.JWTAccStrConfig{
			JWKSUrls:       []string{JWKSUrls},
			TrustedIssuers: []string{trustedIssuers},
			RequiredScopes: []string{"atgo"},
		})
}

func getRawConfig(config any) *runtime.RawExtension {
	bytes, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: bytes,
	}
}
