package gateway

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/version"
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

		It("should configure kyma-gateway-certs and expected secret labels for Gardener path", func() {
			// given
			apiGateway := getApiGateway(true)
			k8sClient := createFakeClient()

			// when
			err := reconcileKymaGatewayCertificate(context.Background(), k8sClient, apiGateway, "some.gardener.domain")

			// then
			Expect(err).ShouldNot(HaveOccurred())

			cert := v1alpha1.Certificate{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertificateName, Namespace: certificateDefaultNamespace}, &cert)).Should(Succeed())
			Expect(*cert.Spec.SecretName).To(Equal(kymaGatewayCertSecretName))
			Expect(*cert.Spec.CommonName).To(Equal("*.some.gardener.domain"))
			Expect(cert.Spec.SecretLabels).To(Equal(map[string]string{
				"kyma-project.io/module":      "api-gateway",
				"app.kubernetes.io/name":      "api-gateway-operator",
				"app.kubernetes.io/instance":  "api-gateway-operator-default",
				"app.kubernetes.io/version":   version.GetModuleVersion(),
				"app.kubernetes.io/component": "operator",
				"app.kubernetes.io/part-of":   "api-gateway",
			}))
		})

	})

	Context("reconcileNonGardenerCertificateSecret", func() {
		It("should create Certificate with default name and namespace", func() {
			// given
			apiGateway := getApiGateway(true)
			k8sClient := createFakeClient()

			// when
			err := reconcileNonGardenerCertificateSecret(context.Background(), k8sClient, apiGateway)

			// then
			Expect(err).ShouldNot(HaveOccurred())

			secret := v1.Secret{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: kymaGatewayCertSecretName, Namespace: certificateDefaultNamespace}, &secret)).Should(Succeed())
			Expect(secret.Data).To(HaveKey("tls.key"))
			Expect(secret.Data).To(HaveKey("tls.crt"))
			Expect(secret.Labels).To(Equal(map[string]string{
				"kyma-project.io/module":      "api-gateway",
				"app.kubernetes.io/name":      "api-gateway-operator",
				"app.kubernetes.io/instance":  "api-gateway-operator-default",
				"app.kubernetes.io/version":   version.GetModuleVersion(),
				"app.kubernetes.io/component": "operator",
				"app.kubernetes.io/part-of":   "api-gateway",
			}))
		})

		It("should not contain certificate that will expire in one month", func() {
			// given
			apiGateway := getApiGateway(true)
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
