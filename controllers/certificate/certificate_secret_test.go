package certificate_test

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	"k8s.io/utils/ptr"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/controllers/certificate"
	. "github.com/onsi/ginkgo/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("InitialiseCertificateSecret", func() {
	It("Should create new secret with certificate if was previously not found", func() {
		// given
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-gateway-controller-manager",
				Namespace: "kyma-system",
			},
		}

		mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "api-gateway-mutating-webhook-configuration",
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "api-gateway-mutating-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "kyma-system",
							Name:      "api-gateway-",
							Path:      ptr.To("/mutate-gateway-kyma-project-io-v1beta1-apirule"),
						},
					},
				},
			},
		}
		validatingWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "api-gateway-validating-webhook-configuration",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "api-gateway-mutating-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "kyma-system",
							Name:      "api-gateway-",
							Path:      ptr.To("/validate-gateway-kyma-project-io-v1beta1-apirule"),
						},
					},
				},
			},
		}

		c := createFakeClient(deployment, mutatingWebhookConfig, validatingWebhookConfig)

		crd := getCRD([]byte{})
		Expect(c.Create(context.Background(), crd)).To(Succeed())

		secret := corev1.Secret{}
		err := c.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &secret)
		Expect(err).Should(HaveOccurred())
		Expect(apierrs.IsNotFound(err)).Should(BeTrue())

		// when
		Expect(certificate.InitialiseCertificateSecret(context.Background(), c, logr.Discard())).Should(Succeed())

		// then
		secret = corev1.Secret{}
		Expect(c.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &secret)).Should(Succeed())

		_, err = cert.ParseCertsPEM(secret.Data["tls.crt"])
		Expect(err).ShouldNot(HaveOccurred())

		crd = &apiextensionsv1.CustomResourceDefinition{}
		Expect(c.Get(ctx, types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, crd)).Should(Succeed())
		Expect(crd.Spec.Conversion.Webhook.ClientConfig.CABundle).To(Equal(secret.Data["tls.crt"]))
	})

	It("Should not create new secret if already existing", func() {
		// given
		cert, key, err := certificate.GenerateSelfSignedCertificate("api-gateway-webhook-service", nil, []string{}, time.Minute*1)
		Expect(err).ShouldNot(HaveOccurred())

		secret := getSecret(cert, key)
		crd := getCRD(cert)
		mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "api-gateway-mutating-webhook-configuration",
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "api-gateway-mutating-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "kyma-system",
							Name:      "api-gateway-",
							Path:      ptr.To("/mutate-gateway-kyma-project-io-v1beta1-apirule"),
						},
					},
				},
			},
		}
		validatingWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "api-gateway-validating-webhook-configuration",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "api-gateway-mutating-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "kyma-system",
							Name:      "api-gateway-",
							Path:      ptr.To("/validate-gateway-kyma-project-io-v1beta1-apirule"),
						},
					},
				},
			},
		}

		c := createFakeClient(secret, mutatingWebhookConfig, validatingWebhookConfig)
		Expect(c.Create(context.Background(), crd)).To(Succeed())

		// when
		Expect(certificate.InitialiseCertificateSecret(context.Background(), c, logr.Discard())).Should(Succeed())

		// then
		currentSecret := corev1.Secret{}
		Expect(c.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &currentSecret)).ShouldNot(HaveOccurred())
		Expect(currentSecret.Data["tls.crt"]).Should(Equal(secret.Data["tls.crt"]))

		crd = &apiextensionsv1.CustomResourceDefinition{}
		Expect(c.Get(ctx, types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, crd)).Should(Succeed())
		Expect(crd.Spec.Conversion.Webhook.ClientConfig.CABundle).To(Equal(secret.Data["tls.crt"]))
	})
})

var _ = Describe("ReadCertificateSecret", func() {
	It("Should be able to read already created certificate and is available for webhook server", func() {
		// given
		cert, key, err := certificate.GenerateSelfSignedCertificate("api-gateway-webhook-service", nil, []string{}, time.Minute*1)
		Expect(err).ShouldNot(HaveOccurred())

		secret := getSecret(cert, key)
		crd := getCRD(cert)

		c := createFakeClient(secret)
		Expect(c.Create(context.Background(), crd)).To(Succeed())

		// when
		Expect(certificate.ReadCertificateSecret(context.Background(), c, logr.Discard())).Should(Succeed())

		// then
		tlsCert, err := certificate.GetCertificate(ptr.To(tls.ClientHelloInfo{}))
		Expect(err).ShouldNot(HaveOccurred())

		certDERBlock, _ := pem.Decode(secret.Data["tls.crt"])
		Expect(certDERBlock.Type).Should(Equal("CERTIFICATE"))
		Expect(certDERBlock.Bytes).Should(Equal(tlsCert.Certificate[0]))

		keyDERBlock, _ := pem.Decode(secret.Data["tls.key"])
		Expect(keyDERBlock.Type).Should(Equal("RSA PRIVATE KEY"))

		privateKey, err := parsePrivateKey(keyDERBlock.Bytes)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(privateKey.(*rsa.PrivateKey).Equal(tlsCert.PrivateKey)).Should(BeTrue())
	})
})

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("tls: found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, errors.New("tls: failed to parse private key")
}
