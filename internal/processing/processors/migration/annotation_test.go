package migration_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
)

var _ = Describe("ApplyMigrationAnnotation", func() {
	It("should add migration step annotation", func() {
		apirule := gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "test-namespace",
				Name:        "test-name",
				Annotations: map[string]string{},
			},
		}

		log := logr.Discard()
		// when
		migration.ApplyMigrationAnnotation(log, &apirule)

		// then
		Expect(apirule.GetAnnotations()).To(HaveKey("gateway.kyma-project.io/migration-step"))
		Expect(apirule.GetAnnotations()["gateway.kyma-project.io/migration-step"]).To(Equal("apply-istio-authorization"))
	})

	It("should remove migration step annotation", func() {
		apirule := gatewayv1beta1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-namespace",
				Name:      "test-name",
				Annotations: map[string]string{
					"gateway.kyma-project.io/migration-step": "vs-switch-to-service",
				},
			},
		}

		log := logr.Discard()

		// when
		migration.ApplyMigrationAnnotation(log, &apirule)

		// then
		Expect(apirule.GetAnnotations()).ToNot(HaveKey("gateway.kyma-project.io/migration-step"))
	})
})
