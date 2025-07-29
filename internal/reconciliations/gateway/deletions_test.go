package gateway

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Deletions", func() {
	Context("reconcileDeletedResources", func() {
		It("should remove ConfigMaps for APIRule UI v1beta1 and v2alpha1", func() {
			// given
			k8sClient := createFakeClient(
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "api-gateway-apirule-v1beta1-ui.operator.kyma-project.io",
						Namespace: "kyma-system",
					},
				},
				&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "api-gateway-apirule-v2alpha1-ui.operator.kyma-project.io",
						Namespace: "kyma-system",
					},
				},
			)

			// when
			err := reconcilev1beta1andv2alpha1UIDeletion(context.Background(), k8sClient)

			// then
			Expect(err).ShouldNot(HaveOccurred())

			Expect(errors.IsNotFound(k8sClient.Get(context.Background(), client.ObjectKey{Name: "api-gateway-apirule-v1beta1-ui.operator.kyma-project.io", Namespace: "kyma-system"}, &v1.ConfigMap{}))).Should(BeTrue())
			Expect(errors.IsNotFound(k8sClient.Get(context.Background(), client.ObjectKey{Name: "api-gateway-apirule-v2alpha1-ui.operator.kyma-project.io", Namespace: "kyma-system"}, &v1.ConfigMap{}))).Should(BeTrue())
		})
	})
})
