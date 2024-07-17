package migration_test

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AnnotationProcessor", func() {
	It("should add migration step annotation", func() {
		apirule := gatewayv2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}
		config := processing.ReconciliationConfig{}
		scheme := runtime.NewScheme()
		err := gatewayv2alpha1.AddToScheme(scheme)
		Expect(err).To(BeNil())

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

		// when
		result, err := migration.NewAnnotationProcessor(config, &apirule).EvaluateReconciliation(context.Background(), k8sClient)

		// then
		Expect(err).To(BeNil())

		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("update"))
		Expect(result[0].Obj.GetAnnotations()).To(HaveKey("gateway.kyma-project.io/migration-step"))
		Expect(result[0].Obj.GetAnnotations()["gateway.kyma-project.io/migration-step"]).To(Equal("apply-istio-authorization"))
	})

	It("should remove migration step annotation", func() {
		apirule := gatewayv2alpha1.APIRule{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"gateway.kyma-project.io/migration-step": "vs-switch-to-service",
				},
			},
		}
		config := processing.ReconciliationConfig{}
		scheme := runtime.NewScheme()
		err := gatewayv2alpha1.AddToScheme(scheme)
		Expect(err).To(BeNil())

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

		// when
		result, err := migration.NewAnnotationProcessor(config, &apirule).EvaluateReconciliation(context.Background(), k8sClient)

		// then
		Expect(err).To(BeNil())

		Expect(result).To(HaveLen(1))
		Expect(result[0].Action.String()).To(Equal("update"))
		Expect(result[0].Obj.GetAnnotations()).ToNot(HaveKey("gateway.kyma-project.io/migration-step"))
	})
})
