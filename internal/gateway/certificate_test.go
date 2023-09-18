package gateway

import (
	"context"
	"github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Certificate", func() {

	Context("reconcileCertificate", func() {

		It("should create Certificate with secret name and domain", func() {
			// given
			k8sClient := createFakeClient()

			// when
			err := reconcileCertificate(context.TODO(), k8sClient, "test", "test-domain.com", "test-cert-secret")

			// then
			Expect(err).ShouldNot(HaveOccurred())

			cert := v1alpha1.Certificate{}
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "istio-system"}, &cert)).Should(Succeed())
			Expect(*cert.Spec.SecretName).To(Equal("test-cert-secret"))
			Expect(*cert.Spec.CommonName).To(Equal("*.test-domain.com"))
		})

		It("should reapply disclaimer annotation on Certificate when it was removed", func() {
			// given
			k8sClient := createFakeClient()
			Expect(reconcileCertificate(context.TODO(), k8sClient, "test", "test-domain.com", "test-cert-secret")).Should(Succeed())

			By("removing disclaimer annotation from certificate")
			cert := v1alpha1.Certificate{}
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "istio-system"}, &cert)).Should(Succeed())
			cert.Annotations = nil
			Expect(k8sClient.Update(context.TODO(), &cert)).Should(Succeed())

			// when
			Expect(reconcileCertificate(context.TODO(), k8sClient, "test", "test-domain.com", "test-cert-secret")).Should(Succeed())

			// then
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: "test", Namespace: "istio-system"}, &cert)).Should(Succeed())

			Expect(cert.Annotations).To(HaveKeyWithValue("apigateways.operator.kyma-project.io/managed-by-disclaimer",
				"DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."))
		})
	})
})
