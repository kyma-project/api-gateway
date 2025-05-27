package conditions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kyma-project/api-gateway/internal/conditions"
)

var _ = Describe("ReasonMessage", func() {
	Context("AdditionalMessage", func() {
		It("should return a condition with custom message", func() {
			// given
			rm := conditions.ReconcileProcessing

			// when
			result := rm.AdditionalMessage(" my-custom-message")

			// then
			Expect(result.Condition().Message).To(Equal("Reconcile processing my-custom-message"))
		})

	})

})
