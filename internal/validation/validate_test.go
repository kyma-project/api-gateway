package validation

import (
	"encoding/json"

	"testing"

	"github.com/kyma-incubator/api-gateway/api/v2alpha1"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ory/oathkeeper-maester/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validators Suite")
}

var _ = Describe("Validate function", func() {

	It("Should fail for empty rules", func() {
		//given
		input := &gatewayv2alpha1.Gate{
			Spec: gatewayv2alpha1.GateSpec{
				Rules: nil,
			},
		}

		//when
		problems := (&Gate{}).Validate(input)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("No rules defined"))
	})

	It("Should detect several problems", func() {
		//given
		input := &gatewayv2alpha1.Gate{
			Spec: gatewayv2alpha1.GateSpec{
				Rules: []gatewayv2alpha1.Rule{
					gatewayv2alpha1.Rule{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("noop", simpleJWTConfig()),
							toAuthenticator("jwt", emptyConfig()),
						},
					},
					gatewayv2alpha1.Rule{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("anonymous", simpleJWTConfig()),
						},
					},
					gatewayv2alpha1.Rule{
						Path: "/def",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("non-existing", nil),
						},
					},
					gatewayv2alpha1.Rule{
						Path:             "/ghi",
						AccessStrategies: []*rulev1alpha1.Authenticator{},
					},
				},
			},
		}
		//when
		problems := (&Gate{}).Validate(input)

		//then
		Expect(problems).To(HaveLen(6))
		Expect(problems[0].AttributePath).To(Equal(".spec.rules"))
		Expect(problems[0].Message).To(Equal("multiple rules defined for the same path"))

		Expect(problems[1].AttributePath).To(Equal(".spec.rules[0].accessStrategies[0].config"))
		Expect(problems[1].Message).To(Equal("strategy: noop does not support configuration"))

		Expect(problems[2].AttributePath).To(Equal(".spec.rules[0].accessStrategies[1].config"))
		Expect(problems[2].Message).To(Equal("supplied config cannot be empty"))

		Expect(problems[3].AttributePath).To(Equal(".spec.rules[1].accessStrategies[0].config"))
		Expect(problems[3].Message).To(Equal("strategy: anonymous does not support configuration"))

		Expect(problems[4].AttributePath).To(Equal(".spec.rules[2].accessStrategies[0].handler"))
		Expect(problems[4].Message).To(Equal("Unsupported accessStrategy: non-existing"))

		Expect(problems[5].AttributePath).To(Equal(".spec.rules[3].accessStrategies"))
		Expect(problems[5].Message).To(Equal("No accessStrategies defined"))
	})

	It("Should succedd for valid input", func() {
		//given
		input := &gatewayv2alpha1.Gate{
			Spec: gatewayv2alpha1.GateSpec{
				Rules: []gatewayv2alpha1.Rule{
					gatewayv2alpha1.Rule{
						Path: "/abc",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("jwt", simpleJWTConfig()),
							toAuthenticator("noop", emptyConfig()),
						},
					},
					gatewayv2alpha1.Rule{
						Path: "/bcd",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("anonymous", emptyConfig()),
						},
					},
					gatewayv2alpha1.Rule{
						Path: "/def",
						AccessStrategies: []*rulev1alpha1.Authenticator{
							toAuthenticator("allow", nil),
						},
					},
				},
			},
		}
		//when
		problems := (&Gate{}).Validate(input)

		//then
		Expect(problems).To(HaveLen(0))
	})
})

var _ = Describe("Validator for", func() {

	Describe("NoConfig access strategy", func() {

		It("Should fail with non-empty config", func() {
			//given
			handler := &v1alpha1.Handler{Name: "noop", Config: simpleJWTConfig("http://atgo.org")}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("strategy: noop does not support configuration"))
		})

		It("Should succeed with empty config: {}", func() {
			//given
			handler := &v1alpha1.Handler{Name: "noop", Config: emptyConfig()}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})

		It("Should succeed with null config", func() {
			//given
			handler := &v1alpha1.Handler{Name: "noop", Config: nil}

			//when
			problems := (&noConfigAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})

	Describe("JWT access strategy", func() {

		It("Should fail with empty config", func() {
			//given
			handler := &v1alpha1.Handler{Name: "jwt", Config: emptyConfig()}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("supplied config cannot be empty"))
		})

		It("Should fail for config with invalid trustedIssuers", func() {
			//given
			handler := &v1alpha1.Handler{Name: "jwt", Config: simpleJWTConfig("a t g o")}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config.trusted_issuers[0]"))
			Expect(problems[0].Message).To(Equal("value is empty or not a valid url"))
		})

		It("Should fail for invalid JSON", func() {
			//given
			handler := &v1alpha1.Handler{Name: "jwt", Config: &runtime.RawExtension{Raw: []byte("/abc]")}}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(1))
			Expect(problems[0].AttributePath).To(Equal("some.attribute.config"))
			Expect(problems[0].Message).To(Equal("Can't read json: invalid character '/' looking for beginning of value"))
		})

		It("Should succeed with valid config", func() {
			//given
			handler := &v1alpha1.Handler{Name: "jwt", Config: simpleJWTConfig()}

			//when
			problems := (&jwtAccStrValidator{}).Validate("some.attribute", handler)

			//then
			Expect(problems).To(HaveLen(0))
		})
	})
})

func emptyConfig() *runtime.RawExtension {
	return getRawConfig(
		&v2alpha1.JWTAccStrConfig{})
}

func simpleJWTConfig(trustedIssuers ...string) *runtime.RawExtension {
	return getRawConfig(
		&v2alpha1.JWTAccStrConfig{
			TrustedIssuers: trustedIssuers,
			RequiredScopes: []string{"atgo"},
		})
}

func getRawConfig(config *v2alpha1.JWTAccStrConfig) *runtime.RawExtension {
	bytes, err := json.Marshal(config)
	Expect(err).To(BeNil())
	return &runtime.RawExtension{
		Raw: bytes,
	}
}

func toAuthenticator(name string, config *runtime.RawExtension) *rulev1alpha1.Authenticator {
	return &rulev1alpha1.Authenticator{
		Handler: &v1alpha1.Handler{
			Name:   name,
			Config: config,
		},
	}
}
