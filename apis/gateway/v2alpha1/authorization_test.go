package v2alpha1_test

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Authorization", func() {

	Describe("JwtAuthorization HasRequiredScopes", func() {

		It("should return false when no required scopes are defined", func() {
			sut := v2alpha1.JwtAuthorization{}
			Expect(sut.HasRequiredScopes()).To(BeFalse())
		})

		It("should return true when required scopes are defined", func() {
			sut := v2alpha1.JwtAuthorization{RequiredScopes: []string{"scope1", "scope2"}}
			Expect(sut.HasRequiredScopes()).To(BeTrue())
		})

		It("should return false when required scopes are defined but empty", func() {
			sut := v2alpha1.JwtAuthorization{RequiredScopes: []string{}}
			Expect(sut.HasRequiredScopes()).To(BeFalse())
		})

	})
})
