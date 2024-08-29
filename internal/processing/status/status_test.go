package status_test

import (
	"errors"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	v2alpha1status "github.com/kyma-project/api-gateway/internal/processing/status"
	"strings"

	"github.com/kyma-project/api-gateway/internal/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status", func() {
	Context("generateValidationStatus", func() {

		f1 := validation.Failure{AttributePath: "name", Message: "is wrong"}
		f2 := validation.Failure{AttributePath: "gateway", Message: "is bad"}
		f3 := validation.Failure{AttributePath: "service.name", Message: "is too short"}
		f4 := validation.Failure{AttributePath: "service.port", Message: "is too big"}
		f5 := validation.Failure{AttributePath: "service.host", Message: "is invalid"}

		It("should generate status for single failure", func() {
			failures := []validation.Failure{f1}
			st := generateValidationStatus(failures)

			Expect(st).NotTo(BeNil())
			Expect(st.Code).To(Equal(gatewayv1beta1.StatusError))
			Expect(st.Description).To(HavePrefix("Validation error: "))
			failureLines := strings.Split(st.Description, "\n")
			Expect(failureLines).To(HaveLen(1))
			Expect(failureLines[0]).To(HaveSuffix("Attribute \"name\": is wrong"))
		})

		It("should generate status for three failures", func() {
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

		It("should generate status for five failures", func() {
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
	Context("v2alpha1", func() {
		Context("GenerateStatusFromFailures", func() {
			It("should generate status for single failure", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{},
				}
				failures := []validation.Failure{
					{
						AttributePath: "name",
						Message:       "is wrong",
					},
				}
				s.GenerateStatusFromFailures(failures)
				Expect(s.ApiRuleStatus.State).To(Equal(gatewayv2alpha1.Error))
				Expect(s.ApiRuleStatus.Description).To(Equal("Validation errors: Attribute 'name': is wrong"))
			})
			It("should generate status for 2 failures", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{},
				}
				failures := []validation.Failure{
					{
						AttributePath: "name",
						Message:       "is wrong",
					},
					{
						AttributePath: "service.name",
						Message:       "is too short",
					},
				}
				s.GenerateStatusFromFailures(failures)
				Expect(s.ApiRuleStatus.State).To(Equal(gatewayv2alpha1.Error))
				Expect(s.ApiRuleStatus.Description).To(Equal("Validation errors: Attribute 'name': is wrong\nAttribute 'service.name': is too short"))
			})
			It("should generate status for 5 failures", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{},
				}
				failures := []validation.Failure{
					{
						AttributePath: "name",
						Message:       "is wrong",
					},
					{
						AttributePath: "service.name",
						Message:       "is too short",
					},
					{
						AttributePath: "service.port",
						Message:       "is too big",
					},
					{
						AttributePath: "service.host",
						Message:       "is invalid",
					},
					{
						AttributePath: "service.host",
						Message:       "is too short",
					},
				}
				s.GenerateStatusFromFailures(failures)
				Expect(s.ApiRuleStatus.State).To(Equal(gatewayv2alpha1.Error))
				Expect(s.ApiRuleStatus.Description).To(Equal("Validation errors: " +
					"Attribute 'name': is wrong\n" +
					"Attribute 'service.name': is too short\n" +
					"Attribute 'service.port': is too big\n" +
					"2 more error(s)..."))
			})
			It("should generate Ready for no failures", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{},
				}
				s.GenerateStatusFromFailures(nil)
				Expect(s.ApiRuleStatus.State).To(Equal(gatewayv2alpha1.Ready))
				Expect(s.ApiRuleStatus.Description).To(Equal("No errors"))
			})
		})
		Context("GetStatusForErrorMap", func() {
			It("should set Error state for errorMap", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{},
				}
				errMap := map[v2alpha1status.ResourceSelector][]error{
					0: {errors.New("one error"), errors.New("another error")},
					1: {errors.New("one error"), errors.New("another error")},
				}
				s.GetStatusForErrorMap(errMap)
				Expect(s.ApiRuleStatus.State).To(Equal(gatewayv2alpha1.Error))
				Expect(s.ApiRuleStatus.Description).To(Equal(
					"ApiRuleErrors: one error, another error\n" +
						"VirtualServiceErrors: one error, another error"))
			})
			It("should set Ready state for no errorMap", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{},
				}
				s.GenerateStatusFromFailures(nil)
				Expect(s.ApiRuleStatus.State).To(Equal(gatewayv2alpha1.Ready))
				Expect(s.ApiRuleStatus.Description).To(Equal("No errors"))
			})
		})
		Context("UpdateStatus", func() {
			It("should return error if v2alpha1 status is empty", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{},
				}
				Expect(s.UpdateStatus(nil)).To(HaveOccurred())
			})
			It("should update status from correct Status", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
						State:       gatewayv2alpha1.Error,
						Description: "some description",
					}}
				expect := &gatewayv2alpha1.APIRuleStatus{}
				Expect(s.UpdateStatus(expect)).ToNot(HaveOccurred())
				Expect(expect.State).To(Equal(gatewayv2alpha1.Error))
				Expect(expect.Description).To(Equal("some description"))

			})
		})
		Context("HasError", func() {
			It("should return true if state equals error", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{
					ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
						State: gatewayv2alpha1.Error,
					},
				}
				Expect(s.HasError()).To(Equal(true))
			})
			It("should return false if APIRuleStatus is nil", func() {
				s := v2alpha1status.ReconciliationV2alpha1Status{}
				Expect(s.HasError()).To(Equal(false))
			})
		})
	})
})

func generateValidationStatus(failures []validation.Failure) *gatewayv1beta1.APIRuleResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, generateValidationDescription(failures))
}

func generateValidationDescription(failures []validation.Failure) string {
	var description string

	if len(failures) == 1 {
		description = "Validation error: "
		description += fmt.Sprintf("Attribute \"%s\": %s", failures[0].AttributePath, failures[0].Message)
	} else {
		const maxEntries = 3
		description = "Multiple validation errors: "
		for i := 0; i < len(failures) && i < maxEntries; i++ {
			description += fmt.Sprintf("\nAttribute \"%s\": %s", failures[i].AttributePath, failures[i].Message)
		}
		if len(failures) > maxEntries {
			description += fmt.Sprintf("\n%d more error(s)...", len(failures)-maxEntries)
		}
	}

	return description
}

func toStatus(c gatewayv1beta1.StatusCode, desc string) *gatewayv1beta1.APIRuleResourceStatus {
	return &gatewayv1beta1.APIRuleResourceStatus{
		Code:        c,
		Description: desc,
	}
}
