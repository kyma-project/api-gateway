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

	})
})
