package v1beta1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
)

var _ = Describe("Mutators", func() {

	Describe("CookieMutatorConfig HasCookies", func() {

		It("should return false when no cookies are defined", func() {
			sut := v1beta1.CookieMutatorConfig{}
			Expect(sut.HasCookies()).To(BeFalse())
		})

		It("should return true when cookies are defined", func() {
			sut := v1beta1.CookieMutatorConfig{Cookies: map[string]string{"cookie1": "value1"}}
			Expect(sut.HasCookies()).To(BeTrue())
		})

		It("should return false when cookies are defined but empty", func() {
			sut := v1beta1.CookieMutatorConfig{Cookies: map[string]string{}}
			Expect(sut.HasCookies()).To(BeFalse())
		})
	})

	Describe("CookieMutatorConfig ToString", func() {

		It("should return empty string when no cookies are defined", func() {
			sut := v1beta1.CookieMutatorConfig{}
			Expect(sut.ToString()).To(Equal(""))
		})

		It("should return string with cookies", func() {
			sut := v1beta1.CookieMutatorConfig{Cookies: map[string]string{"cookie1": "value1", "cookie2": "value2"}}
			Expect(sut.ToString()).To(ContainSubstring("cookie1=value1"))
			Expect(sut.ToString()).To(ContainSubstring("; "))
			Expect(sut.ToString()).To(ContainSubstring("cookie2=value2"))
		})

		It("should return empty string when cookies are defined but empty", func() {
			sut := v1beta1.CookieMutatorConfig{Cookies: map[string]string{}}
			Expect(sut.ToString()).To(Equal(""))
		})
	})

	Describe("HeaderMutatorConfig HasHeaders", func() {

		It("should return false when no headers are defined", func() {
			sut := v1beta1.HeaderMutatorConfig{}
			Expect(sut.HasHeaders()).To(BeFalse())
		})

		It("should return true when headers are defined", func() {
			sut := v1beta1.HeaderMutatorConfig{Headers: map[string]string{"header1": "value1"}}
			Expect(sut.HasHeaders()).To(BeTrue())
		})

		It("should return false when headers are defined but empty", func() {
			sut := v1beta1.HeaderMutatorConfig{Headers: map[string]string{}}
			Expect(sut.HasHeaders()).To(BeFalse())
		})
	})
})
