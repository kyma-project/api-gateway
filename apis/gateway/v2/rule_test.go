package v2_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/apis/gateway/v2"
)

var _ = Describe("Rule", func() {

	Describe("ContainsAccessStrategy", func() {

		It("should return false when noAuth does not exist in the object", func() {
			rule := v2.Rule{
				Jwt: &v2.JwtConfig{},
			}

			Expect(rule.ContainsNoAuth()).To(BeFalse())
		})

		It("should return true when noAuth exists in the object", func() {
			value := true

			rule := v2.Rule{
				NoAuth: &value,
			}

			Expect(rule.ContainsNoAuth()).To(BeTrue())
		})

		It("should return true when JWT exists in the object", func() {
			rule := v2.Rule{
				Jwt: &v2.JwtConfig{},
			}

			Expect(rule.ContainsAccessStrategyJwt()).To(BeTrue())
		})

		It("should return false when no access strategy is in the object", func() {
			rule := v2.Rule{}

			Expect(rule.ContainsAccessStrategyJwt()).To(BeFalse())
			Expect(rule.ContainsNoAuth()).To(BeFalse())
		})

	})

	Describe("ConvertHttpMethodsToStrings", func() {

		It("should convert the HttpMethod slice to a string slice", func() {
			methods := []v2.HttpMethod{
				http.MethodGet,
				http.MethodPost,
			}

			expected := []string{
				"GET",
				"POST",
			}

			Expect(v2.ConvertHttpMethodsToStrings(methods)).To(ConsistOf(expected))
		})

		It("should return an empty slice when the input slice is nil", func() {
			Expect(v2.ConvertHttpMethodsToStrings(nil)).To(BeEmpty())
		})

	})

	Describe("AppliesToAllPaths", func() {

		It("should return true when the path applies to all paths", func() {
			rule := v2.Rule{
				Path: "/*",
			}

			Expect(rule.AppliesToAllPaths()).To(BeTrue())
		})

		It("should return false when the path does not apply to all paths", func() {
			rule := v2.Rule{
				Path: "/",
			}

			Expect(rule.AppliesToAllPaths()).To(BeFalse())
		})
	})
})
