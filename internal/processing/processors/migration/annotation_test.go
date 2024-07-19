package migration_test

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

		scheme := runtime.NewScheme()
		err := gatewayv2alpha1.AddToScheme(scheme)
		Expect(err).To(BeNil())
		err = gatewayv1beta1.AddToScheme(scheme)
		Expect(err).To(BeNil())

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&apirule).Build()
		log := logr.Discard()
		// when
		err = migration.ApplyMigrationAnnotation(context.Background(), k8sClient, &log, &apirule)
		Expect(err).To(BeNil())

		// then
		err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-name", Namespace: "test-namespace"}, &apirule)
		Expect(err).To(BeNil())
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

		scheme := runtime.NewScheme()
		err := gatewayv2alpha1.AddToScheme(scheme)
		Expect(err).To(BeNil())
		err = gatewayv1beta1.AddToScheme(scheme)
		Expect(err).To(BeNil())

		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&apirule).Build()
		log := logr.Discard()

		// when
		err = migration.ApplyMigrationAnnotation(context.Background(), k8sClient, &log, &apirule)
		Expect(err).To(BeNil())

		// then
		err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-name", Namespace: "test-namespace"}, &apirule)
		Expect(err).To(BeNil())
		Expect(apirule.GetAnnotations()).ToNot(HaveKey("gateway.kyma-project.io/migration-step"))
	})
})
