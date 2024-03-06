package conditions_test

import (
	"github.com/kyma-project/api-gateway/internal/conditions"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReasonMessage", func() {
	Context("AdditionalMessage", func() {
		It("should extend existing condition message with string", func() {
			// given
			rm := conditions.ReconcileProcessing

			// when
			rm.AdditionalMessage(" my-custom-message")

			// then
			Expect(rm.Condition().Message).To(Equal("Reconcile processing my-custom-message"))
		})

	})

})
