package v2alpha1_test

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rule", func() {

	Describe("ContainsAccessStrategy", func() {

		It("should return false when noAuth does not exist in the object", func() {
			rule := v2alpha1.Rule{
				Jwt: &v2alpha1.JwtConfig{},
			}

			Expect(rule.ContainsNoAuth()).To(BeFalse())
		})

		It("should return true when noAuth exists in the object", func() {
			value := true

			rule := v2alpha1.Rule{
				NoAuth: &value,
			}

			Expect(rule.ContainsNoAuth()).To(BeTrue())
		})

		It("should return true when JWT exists in the object", func() {
			rule := v2alpha1.Rule{
				Jwt: &v2alpha1.JwtConfig{},
			}

			Expect(rule.ContainsAccessStrategyJwt()).To(BeTrue())
		})

		It("should return false when no access strategy is in the object", func() {
			rule := v2alpha1.Rule{}

			Expect(rule.ContainsAccessStrategyJwt()).To(BeFalse())
			Expect(rule.ContainsNoAuth()).To(BeFalse())
		})

	})

	Describe("ConvertHttpMethodsToStrings", func() {

		It("should convert the HttpMethod slice to a string slice", func() {
			methods := []v2alpha1.HttpMethod{
				http.MethodGet,
				http.MethodPost,
			}

			expected := []string{
				"GET",
				"POST",
			}

			Expect(v2alpha1.ConvertHttpMethodsToStrings(methods)).To(ConsistOf(expected))
		})

		It("should return an empty slice when the input slice is nil", func() {
			Expect(v2alpha1.ConvertHttpMethodsToStrings(nil)).To(BeEmpty())
		})

	})

	Describe("AppliesToAllPaths", func() {

		It("should return true when the path applies to all paths", func() {
			rule := v2alpha1.Rule{
				Path: "/*",
			}

			Expect(rule.AppliesToAllPaths()).To(BeTrue())
		})

		It("should return false when the path does not apply to all paths", func() {
			rule := v2alpha1.Rule{
				Path: "/",
			}

			Expect(rule.AppliesToAllPaths()).To(BeFalse())
		})
	})
})
