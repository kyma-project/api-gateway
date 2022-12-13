package istio

import (
	"encoding/json"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	istiojwt "github.com/kyma-incubator/api-gateway/internal/types/istio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("JWT Validator", func() {
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
})

func emptyJWTIstioConfig() *runtime.RawExtension {
	return getRawConfig(
		&istiojwt.JwtConfig{})
}

func simpleJWTIstioConfig(trustedIssuers ...string) *runtime.RawExtension {
	issuers := []istiojwt.JwtAuth{}
	for _, issuer := range trustedIssuers {
		issuers = append(issuers, istiojwt.JwtAuth{
			Issuer:  issuer,
			JwksUri: issuer,
		})
	}
	jwtConfig := istiojwt.JwtConfig{Authentications: issuers}
	return getRawConfig(jwtConfig)
}

func testURLJWTIstioConfig(JWKSUrl string, trustedIssuer string) *runtime.RawExtension {
	return getRawConfig(
		istiojwt.JwtConfig{
			Authentications: []istiojwt.JwtAuth{
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
