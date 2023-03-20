package istio

import (
	"github.com/kyma-project/api-gateway/api/v1beta1"
	processingtest "github.com/kyma-project/api-gateway/internal/processing/internal/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mutators validator", func() {
	It("Should fail for handler that is not supported", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "unsupported",
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler"))
		Expect(problems[0].Message).To(Equal("unsupported handler: unsupported"))
	})

	It("Should fail for empty handler", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler"))
		Expect(problems[0].Message).To(Equal("handler cannot be empty"))
	})

	It("Should fail for header handler without config", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("headers cannot be empty"))
	})

	It("Should fail for header handler without headers", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: processingtest.GetRawConfig(
							v1beta1.HeaderMutatorConfig{}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("headers cannot be empty"))
	})

	It("Should fail for header handler with empty headers", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: processingtest.GetRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("headers cannot be empty"))
	})

	It("Should fail for header handler without header name", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: processingtest.GetRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{
									"": "test",
								},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config.headers.name"))
		Expect(problems[0].Message).To(Equal("cannot be empty"))
	})

	It("Should fail for header handler without header value", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: processingtest.GetRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{
									"x-test-header": "",
								},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config.headers.[x-test-header]"))
		Expect(problems[0].Message).To(Equal("header value cannot be empty"))
	})

	It("Should have no failures for header handler with headers", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "header",
						Config: processingtest.GetRawConfig(
							v1beta1.HeaderMutatorConfig{
								Headers: map[string]string{
									"x-test-header": "test",
								},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should fail for cookie handler without config", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("cookies cannot be empty"))
	})

	It("Should fail for cookie handler without cookies", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: processingtest.GetRawConfig(
							v1beta1.CookieMutatorConfig{}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("cookies cannot be empty"))
	})

	It("Should fail for cookie handler with empty cookies", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: processingtest.GetRawConfig(
							v1beta1.CookieMutatorConfig{
								Cookies: map[string]string{},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config"))
		Expect(problems[0].Message).To(Equal("cookies cannot be empty"))
	})

	It("Should fail for cookie handler without cookie name", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: processingtest.GetRawConfig(
							v1beta1.CookieMutatorConfig{
								Cookies: map[string]string{
									"": "test",
								},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config.cookies.name"))
		Expect(problems[0].Message).To(Equal("cannot be empty"))
	})

	It("Should fail for cookie handler without cookie value", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: processingtest.GetRawConfig(
							v1beta1.CookieMutatorConfig{
								Cookies: map[string]string{
									"x-test-cookie": "",
								},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[0].handler.config.cookies.[x-test-cookie]"))
		Expect(problems[0].Message).To(Equal("cookie value cannot be empty"))
	})

	It("Should have no failures for cookie handler with cookies", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: processingtest.GetRawConfig(
							v1beta1.CookieMutatorConfig{
								Cookies: map[string]string{
									"x-test-cookie": "test",
								},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(0))
	})

	It("Should add failures for multiple mutators", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "unsupported",
					},
				},
				{
					Handler: &v1beta1.Handler{
						Name: "",
					},
				},
				{
					Handler: &v1beta1.Handler{
						Name: "totally unsupported",
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(3))
	})

	It("Should fail for duplicated handlers", func() {
		//given
		rule := v1beta1.Rule{
			Mutators: []*v1beta1.Mutator{
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: processingtest.GetRawConfig(
							v1beta1.CookieMutatorConfig{
								Cookies: map[string]string{
									"x-test-cookie": "test",
								},
							}),
					},
				},
				{
					Handler: &v1beta1.Handler{
						Name: "cookie",
						Config: processingtest.GetRawConfig(
							v1beta1.CookieMutatorConfig{
								Cookies: map[string]string{
									"other-cookie": "test",
								},
							}),
					},
				},
			},
		}

		//when
		problems := mutatorsValidator{}.Validate("some.attribute", rule)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].AttributePath).To(Equal("some.attribute.mutators[1].handler.cookie"))
		Expect(problems[0].Message).To(Equal("mutator for same handler already exists"))
	})
})
