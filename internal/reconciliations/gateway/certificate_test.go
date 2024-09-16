package gateway

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Certificate", func() {
	Context("reconcileCertificate", func() {
		It("should create Certificate with secret name and domain", func() {
			// given
			k8sClient := createFakeClient()

			// when
			err := reconcileCertificate(context.Background(), k8sClient, "test", "test-domain.com", "test-cert-secret")

			// then
			Expect(err).ShouldNot(HaveOccurred())

			cert := v1alpha1.Certificate{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: "test", Namespace: certificateDefaultNamespace}, &cert)).Should(Succeed())
			Expect(*cert.Spec.SecretName).To(Equal("test-cert-secret"))
			Expect(*cert.Spec.CommonName).To(Equal("*.test-domain.com"))
		})

	})

	Context("reconcileNonGardenerCertificateSecret", func() {
		It("should create Certificate with default name and namespace", func() {
			// given
			apiGateway := getApiGateway(true, false)
			k8sClient := createFakeClient()

			// when
			err := reconcileNonGardenerCertificateSecret(context.Background(), k8sClient, apiGateway)

			// then
			Expect(err).ShouldNot(HaveOccurred())

			secret := v1.Secret{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertSecretName, Namespace: certificateDefaultNamespace}, &secret)).Should(Succeed())
			Expect(secret.Data).To(HaveKey("tls.key"))
			Expect(secret.Data).To(HaveKey("tls.crt"))
		})

		It("should not contain certificate that will expire in one month", func() {
			// given
			apiGateway := getApiGateway(true, false)
			k8sClient := createFakeClient()

			// when
			err := reconcileNonGardenerCertificateSecret(context.Background(), k8sClient, apiGateway)

			// then
			Expect(err).ShouldNot(HaveOccurred())

			secret := v1.Secret{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertSecretName, Namespace: certificateDefaultNamespace}, &secret)).Should(Succeed())
			Expect(secret.Data).To(HaveKey("tls.crt"))
			willExpireInOneMonth, err := certificateExpireInOneMonth(string(secret.Data["tls.crt"]))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(willExpireInOneMonth).To(BeFalse())
		})
	})
})

func certificateExpireInOneMonth(certPEM string) (bool, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return false, errors.New("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}

	plusOneMonth := time.Now().AddDate(0, 1, 0)
	return plusOneMonth.After(cert.NotAfter), nil
}
