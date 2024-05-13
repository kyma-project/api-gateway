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

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	"k8s.io/utils/ptr"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/controllers/certificate"
	. "github.com/onsi/ginkgo/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Certificate initialisation functions on manager start", Ordered, func() {
	It("Should create new secret with certificate if was previously not found", func() {
		// given
		secret := corev1.Secret{}
		err := k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &secret)
		Expect(err).Should(HaveOccurred())
		Expect(apierrs.IsNotFound(err)).Should(BeTrue())

		// when
		Expect(certificate.InitialiseCertificateSecret(context.Background(), k8sClient, logr.Discard())).Should(Succeed())

		// then
		secret = corev1.Secret{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &secret)).Should(Succeed())
		_, err = cert.ParseCertsPEM(secret.Data["tls.crt"])
		Expect(err).ShouldNot(HaveOccurred())

		crd := apiextensionsv1.CustomResourceDefinition{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "apirules.gateway.kyma-project.io"}, &crd)).Should(Succeed())
		Expect(crd.Spec.Conversion.Webhook.ClientConfig.CABundle).To(Equal(secret.Data["tls.crt"]))
	})

	It("Should not create new secret if already created", func() {
		// given
		originalSecret := corev1.Secret{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &originalSecret)).ShouldNot(HaveOccurred())

		// when
		Expect(certificate.InitialiseCertificateSecret(context.Background(), k8sClient, logr.Discard())).Should(Succeed())

		// then
		currentSecret := corev1.Secret{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &currentSecret)).ShouldNot(HaveOccurred())
		Expect(currentSecret.Data["tls.crt"]).Should(Equal(originalSecret.Data["tls.crt"]))
	})

	It("Should be able to read already created certificate and is available for webhook server", func() {
		// given
		secret := corev1.Secret{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "kyma-system", Name: "api-gateway-webhook-certificate"}, &secret)).ShouldNot(HaveOccurred())

		// when
		Expect(certificate.ReadCertificateSecret(context.Background(), k8sClient, logr.Discard())).Should(Succeed())

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
