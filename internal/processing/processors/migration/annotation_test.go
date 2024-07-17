package migration

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AnnotationProcessor", func() {
	DescribeTable("should return correct next step", func(annotation string, expectedStep string) {
		// when
		step := nextMigrationStep(&gatewayv2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					annotationName: annotation,
				},
			},
		})
		// then
		Expect(string(step)).To(Equal(expectedStep))
	},
		Entry("should return applyIstioAuthorizationMigrationStep when annotation is not set", "", "apply-istio-authorization"),
		Entry("should return switchVsToService when current step is applyIstioAuthorizationMigrationStep", "apply-istio-authorization", "vs-switch-to-service"),
		Entry("should return removeOryRule when current step is switchVsToService", "vs-switch-to-service", "remove-ory-rule"),
		Entry("should return finished when current step is removeOryRule", "remove-ory-rule", "migration-finished"),
	)
})
