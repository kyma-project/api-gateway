package istio

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	processingtest "github.com/kyma-project/api-gateway/internal/processing/processing_test"
)

var _ = Describe("Mutators validator", func() {

	jwtAccessStrategy := []*v1beta1.Authenticator{
		{
			Handler: &v1beta1.Handler{
				Name: "jwt",
			},
		},
	}

	createJwtHandlerRule := func(mutators ...*v1beta1.Mutator) v1beta1.Rule {
		return v1beta1.Rule{
			Mutators:         mutators,
			AccessStrategies: jwtAccessStrategy,
		}
	}

	It("Should fail for handler that is not supported", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "unsupported",
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler"))
		Expect(problems[0].Message).To(Equal("unsupported mutator: unsupported"))
	})

	It("Should fail for empty handler", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler"))
		Expect(problems[0].Message).To(Equal("mutator handler cannot be empty"))
	})

	It("Should fail for header handler without config", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "header",
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("headers cannot be empty"))
	})

	It("Should fail for header handler without headers", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "header",
				Config: processingtest.GetRawConfig(
					v1beta1.HeaderMutatorConfig{}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("headers cannot be empty"))
	})

	It("Should fail for header handler with empty headers", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "header",
				Config: processingtest.GetRawConfig(
					v1beta1.HeaderMutatorConfig{
						Headers: map[string]string{},
					}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("headers cannot be empty"))
	})

	It("Should fail for header handler without header name", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "header",
				Config: processingtest.GetRawConfig(
					v1beta1.HeaderMutatorConfig{
						Headers: map[string]string{
							"": "test",
						},
					}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config.headers.name"))
		Expect(problems[0].Message).To(Equal("cannot be empty"))
	})

	It("Should have no failures for header handler with headers", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "header",
				Config: processingtest.GetRawConfig(
					v1beta1.HeaderMutatorConfig{
						Headers: map[string]string{
							"x-test-header": "test",
						},
					}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for cookie handler without config", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("cookies cannot be empty"))
	})

	It("Should fail for cookie handler without cookies", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
				Config: processingtest.GetRawConfig(
					v1beta1.CookieMutatorConfig{}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("cookies cannot be empty"))
	})

	It("Should fail for cookie handler with empty cookies", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
				Config: processingtest.GetRawConfig(
					v1beta1.CookieMutatorConfig{
						Cookies: map[string]string{},
					}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("cookies cannot be empty"))
	})

	It("Should fail for cookie handler without cookie name", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
				Config: processingtest.GetRawConfig(
					v1beta1.CookieMutatorConfig{
						Cookies: map[string]string{
							"": "test",
						},
					}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config.cookies.name"))
		Expect(problems[0].Message).To(Equal("cannot be empty"))
	})

	It("Should have no failures for cookie handler with cookies", func() {
		//given
		mutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
				Config: processingtest.GetRawConfig(
					v1beta1.CookieMutatorConfig{
						Cookies: map[string]string{
							"x-test-cookie": "test",
						},
					}),
			},
		}

		rule := createJwtHandlerRule(&mutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should add failures for multiple mutators", func() {
		//given
		unsupportedMutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "unsupported",
			},
		}

		noNameMutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "",
			},
		}

		noConfigMutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
			},
		}

		rule := createJwtHandlerRule(&unsupportedMutator, &noNameMutator, &noConfigMutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(3))
	})

	It("Should fail for duplicated handlers", func() {
		//given

		cookieMutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
				Config: processingtest.GetRawConfig(
					v1beta1.CookieMutatorConfig{
						Cookies: map[string]string{
							"x-test-cookie": "test",
						},
					}),
			},
		}
		anotherCookieMutator := v1beta1.Mutator{
			Handler: &v1beta1.Handler{
				Name: "cookie",
				Config: processingtest.GetRawConfig(
					v1beta1.CookieMutatorConfig{
						Cookies: map[string]string{
							"other-cookie": "test",
						},
					}),
			},
		}

		rule := createJwtHandlerRule(&cookieMutator, &anotherCookieMutator)

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[1].handler.cookie"))
		Expect(problems[0].Message).To(Equal("mutator for same handler already exists"))
	})
})
