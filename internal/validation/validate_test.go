package validation

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/internal/helpers"
)

var _ = Describe("ValidateConfig function", func() {
	It("Should fail for missing config", func() {
		//when
		problems := ValidateConfig(nil)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Configuration is missing"))
	})

	It("Should fail for wrong config", func() {
		//given
		input := &helpers.Config{JWTHandler: "foo"}

		//when
		problems := ValidateConfig(input)

		//then
		Expect(problems).To(HaveLen(1))
		Expect(problems[0].Message).To(Equal("Unsupported JWT Handler: foo"))
	})
})
