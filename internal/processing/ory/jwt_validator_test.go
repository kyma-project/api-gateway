package ory

import (
	"encoding/json"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	"testing"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestOryJwtValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Ory JWT Validator Suite",
		[]Reporter{printer.NewProwReporter("api-gateway-ory-jwt-validator-testsuite")})
}

var _ = Describe("Validator for", func() {

	Describe("JWT access strategy", func() {

		It("Should fail with empty config", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: emptyConfig()}

			//when
			problems := (&jwtValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("supplied config cannot be empty"))
		})

		It("Should fail for config with invalid trustedIssuers and JWKSUrls", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: simpleJWTConfig("a t g o")}

			//when
			problems := (&jwtValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(2))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.trusted_issuers[0]"))
			Expect(problems[0].Message).To(ContainSubstring("value is empty or not a valid url"))
			Expect(problems[1].AttributePath).To(Equal("some.attribute.config.jwks_urls[0]"))
			Expect(problems[1].Message).To(ContainSubstring("value is empty or not a valid url"))
		})

		It("Should fail for config with plain HTTP JWKSUrls and trustedIssuers", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("http://issuer.test/.well-known/jwks.json", "http://issuer.test/")}

			//when
			problems := (&jwtValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(2))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.trusted_issuers[0]"))
			Expect(problems[0].Message).To(ContainSubstring("value is not a secured url"))
			Expect(problems[1].AttributePath).To(Equal("some.attribute.config.jwks_urls[0]"))
			Expect(problems[1].Message).To(ContainSubstring("value is not a secured url"))
		})

		It("Should succeed for config with file JWKSUrls and HTTPS trustedIssuers", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("file://.well-known/jwks.json", "https://issuer.test/")}

			//when
			problems := (&jwtValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should succeed for config with HTTPS JWKSUrls and trustedIssuers", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: testURLJWTConfig("https://issuer.test/.well-known/jwks.json", "https://issuer.test/")}

			//when
			problems := (&jwtValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should fail for invalid JSON", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: &runtime.RawExtension{Raw: []byte("/abc]")}}

			//when
			problems := (&jwtValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("Can't read json: invalid character '/' looking for beginning of value"))
		})

		It("Should succeed with valid config", func() {
			//given
			handler := &gatewayv1beta1.Handler{Name: "jwt", Config: simpleJWTConfig()}

			//when
			problems := (&jwtValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})
})

func emptyConfig() *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1beta1.JWTAccStrConfig{})
}

func simpleJWTConfig(trustedIssuers ...string) *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1beta1.JWTAccStrConfig{
			JWKSUrls:       trustedIssuers,
			TrustedIssuers: trustedIssuers,
			RequiredScopes: []string{"atgo"},
		})
}

func testURLJWTConfig(JWKSUrls string, trustedIssuers string) *runtime.RawExtension {
	return getRawConfig(
		&gatewayv1beta1.JWTAccStrConfig{
			JWKSUrls:       []string{JWKSUrls},
			TrustedIssuers: []string{trustedIssuers},
			RequiredScopes: []string{"atgo"},
		})
}

func getRawConfig(config *gatewayv1beta1.JWTAccStrConfig) *runtime.RawExtension {
	bytes, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: bytes,
	}
}
