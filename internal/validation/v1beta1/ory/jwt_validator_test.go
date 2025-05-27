package ory

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/types/ory"
)

var _ = Describe("JWT Validator", func() {
	It("Should fail with empty config", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: emptyConfig()}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
		Expect(problems[0].Message).To(Equal("supplied config cannot be empty"))
	})

	It("Should fail for config with invalid trustedIssuers and JWKSUrls", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: simpleJWTConfig("a t g o")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(2))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.config.trusted_issuers[0]"))
		Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid uri"))
		Expect(problems[1].AttributePath).To(Equal("some.attribute.config.jwks_urls[0]"))
		Expect(problems[1].Message).To(ContainSubstring("value is empty or not a valid uri"))
	})

	It("Should succeed for config with plain HTTP JWKSUrls and trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("http://issuer.test/.well-known/jwks.json", "http://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for config with file JWKSUrls and HTTPS trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("file://.well-known/jwks.json", "https://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should succeed for config with HTTPS JWKSUrls and trustedIssuers", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("https://issuer.test/.well-known/jwks.json", "https://issuer.test/")}

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

	It("Should succeed with valid config", func() {
		//given
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: simpleJWTConfig()}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for config with Istio JWT configuration", func() {
		handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTIstioConfig("https://issuer.test/.well-known/jwks.json", "https://issuer.test/")}

		//when
		problems := (&handlerValidator{}).Validate("some.attribute", handler)

		//then
		Expect(problems).To(Not(BeEmpty()))
	})
})

func emptyConfig() *runtime.RawExtension {
	return getRawConfig(
		&ory.JWTAccStrConfig{})
}

func simpleJWTConfig(trustedIssuers ...string) *runtime.RawExtension {
	return getRawConfig(
		&ory.JWTAccStrConfig{
			JWKSUrls:       trustedIssuers,
			TrustedIssuers: trustedIssuers,
			RequiredScopes: []string{"atgo"},
		})
}

func testURLJWTConfig(JWKSUrls string, trustedIssuers string) *runtime.RawExtension {
	return getRawConfig(
		&ory.JWTAccStrConfig{
			JWKSUrls:       []string{JWKSUrls},
			TrustedIssuers: []string{trustedIssuers},
			RequiredScopes: []string{"atgo"},
		})
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

func getRawConfig(config any) *runtime.RawExtension {
	bytes, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: bytes,
	}
}
