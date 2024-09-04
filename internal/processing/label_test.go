package processing_test

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GetOwnerLabels", func() {
	expectedOwnerLabelKey := "apirule.gateway.kyma-project.io/v1beta1"

	Context("v1beta1", func() {
		It("should return v1beta1 owner label", func() {
			apiRule := gatewayv1beta1.APIRule{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-apirule-psdh34",
					Namespace: "test-namespace",
				},
			}

			labels := processing.GetOwnerLabels(&apiRule)
			Expect(labels).To(HaveKeyWithValue(expectedOwnerLabelKey, "test-apirule-psdh34.test-namespace"))
		})
	})

	Context("v2alpha1", func() {
		It("should return v1beta1 owner label", func() {
			apiRule := gatewayv2alpha1.APIRule{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-apirule-psdh34",
					Namespace: "test-namespace",
				},
			}

			labels := processing.GetOwnerLabelsV2alpha1(&apiRule)
			Expect(labels).To(HaveKeyWithValue(expectedOwnerLabelKey, "test-apirule-psdh34.test-namespace"))
		})
	})
})
