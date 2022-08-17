package controllers

import (
	"strings"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller", func() {
	Describe("generateValidationProblemStatus", func() {

		f1 := validation.Failure{AttributePath: "name", Message: "is wrong"}
		f2 := validation.Failure{AttributePath: "gateway", Message: "is bad"}
		f3 := validation.Failure{AttributePath: "service.name", Message: "is too short"}
		f4 := validation.Failure{AttributePath: "service.port", Message: "is too big"}
		f5 := validation.Failure{AttributePath: "service.host", Message: "is invalid"}

		It("should genereate status for single failure", func() {
			failures := []validation.Failure{f1}
			st := generateValidationStatus(failures)

			Expect(st).NotTo(BeNil())
			Expect(st.Code).To(Equal(gatewayv1beta1.StatusError))
			Expect(st.Description).To(HavePrefix("Validation error: "))
			failureLines := strings.Split(st.Description, "\n")
			Expect(failureLines).To(HaveLen(1))
			Expect(failureLines[0]).To(HaveSuffix("Attribute \"name\": is wrong"))
		})

		It("should genereate status for three failures", func() {
			failures := []validation.Failure{f1, f2, f3}
			st := generateValidationStatus(failures)

			Expect(st).NotTo(BeNil())
			Expect(st.Code).To(Equal(gatewayv1beta1.StatusError))
			failureLines := strings.Split(st.Description, "\n")
			Expect(failureLines).To(HaveLen(4))
			Expect(failureLines[0]).To(Equal("Multiple validation errors: "))
			Expect(failureLines[1]).To(Equal("Attribute \"name\": is wrong"))
			Expect(failureLines[2]).To(Equal("Attribute \"gateway\": is bad"))
			Expect(failureLines[3]).To(Equal("Attribute \"service.name\": is too short"))
		})

		It("should genereate status for five failures", func() {
			failures := []validation.Failure{f1, f2, f3, f4, f5}
			st := generateValidationStatus(failures)

			Expect(st).NotTo(BeNil())
			Expect(st.Code).To(Equal(gatewayv1beta1.StatusError))
			failureLines := strings.Split(st.Description, "\n")
			Expect(failureLines).To(HaveLen(5))
			Expect(failureLines[0]).To(Equal("Multiple validation errors: "))
			Expect(failureLines[1]).To(Equal("Attribute \"name\": is wrong"))
			Expect(failureLines[2]).To(Equal("Attribute \"gateway\": is bad"))
			Expect(failureLines[3]).To(Equal("Attribute \"service.name\": is too short"))
			Expect(failureLines[4]).To(Equal("2 more error(s)..."))
		})
	})
})
