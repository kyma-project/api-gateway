package extgateway

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
)

// CertBundle holds a CA and a leaf certificate/key pair for mTLS testing.
type CertBundle struct {
	CACertPEM     []byte
	ClientCertPEM []byte
	ClientKeyPEM  []byte
	// Subject is the full DN of the client certificate in Envoy/Lua comparison format.
	Subject string
}

var subjectAttributeNames = map[string]string{
	"2.5.4.3":  "CN",
	"2.5.4.6":  "C",
	"2.5.4.7":  "L",
	"2.5.4.8":  "ST",
	"2.5.4.9":  "STREET",
	"2.5.4.10": "O",
	"2.5.4.11": "OU",
	"2.5.4.17": "POSTALCODE",
}

func subjectForEnvoyComparison(cert *x509.Certificate) (string, error) {
	var rdnSequence pkix.RDNSequence
	if _, err := asn1.Unmarshal(cert.RawSubject, &rdnSequence); err != nil {
		return "", fmt.Errorf("unmarshalling raw subject: %w", err)
	}

	parts := make([]string, 0, len(rdnSequence))
	for i := len(rdnSequence) - 1; i >= 0; i-- {
		for _, attribute := range rdnSequence[i] {
			name := subjectAttributeNames[attribute.Type.String()]
			if name == "" {
				name = attribute.Type.String()
			}

			parts = append(parts, fmt.Sprintf("%s=%v", name, attribute.Value))
		}
	}

	return strings.Join(parts, ","), nil
}

// GenerateMTLSCerts generates a CA and a leaf client certificate, then derives
// the exact subject string format used by Envoy/Lua from the generated cert.
func GenerateMTLSCerts(t *testing.T) (CertBundle, error) {
	t.Helper()
	t.Log("Generating mTLS cert bundle")

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return CertBundle{}, fmt.Errorf("generating CA key: %w", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return CertBundle{}, fmt.Errorf("creating CA cert: %w", err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return CertBundle{}, fmt.Errorf("parsing CA cert: %w", err)
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return CertBundle{}, fmt.Errorf("generating client key: %w", err)
	}

	// Keep the subject fields in standard X.509 order; the exact comparison string
	// is derived later from the generated certificate's raw subject.
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Country:            []string{"US"},
			Organization:       []string{"Test Org"},
			OrganizationalUnit: []string{"test-clients"},
			Locality:           []string{"gateway"},
			CommonName:         "test-client/test-region",
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return CertBundle{}, fmt.Errorf("creating client cert: %w", err)
	}
	clientCert, err := x509.ParseCertificate(clientDER)
	if err != nil {
		return CertBundle{}, fmt.Errorf("parsing client cert: %w", err)
	}
	clientSubject, err := subjectForEnvoyComparison(clientCert)
	if err != nil {
		return CertBundle{}, fmt.Errorf("building client cert subject for envoy: %w", err)
	}

	clientKeyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		return CertBundle{}, fmt.Errorf("marshalling client key: %w", err)
	}

	return CertBundle{
		CACertPEM:     pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}),
		ClientCertPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER}),
		ClientKeyPEM:  pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER}),
		Subject:       clientSubject,
	}, nil
}

// GenerateUntrustedClientCert generates a client certificate signed by a different CA.
// Presenting this cert to the mTLS endpoint will cause the TLS handshake to fail.
func GenerateUntrustedClientCert(t *testing.T) (clientCertPEM, clientKeyPEM []byte, err error) {
	t.Helper()
	t.Log("Generating untrusted client cert")

	untrustedCA, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generating untrusted CA key: %w", err)
	}
	untrustedCATemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(10),
		Subject:               pkix.Name{CommonName: "Untrusted CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	untrustedCACERTDER, err := x509.CreateCertificate(rand.Reader, untrustedCATemplate, untrustedCATemplate, &untrustedCA.PublicKey, untrustedCA)
	if err != nil {
		return nil, nil, fmt.Errorf("creating untrusted CA cert: %w", err)
	}
	untrustedCACert, err := x509.ParseCertificate(untrustedCACERTDER)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing untrusted CA cert: %w", err)
	}

	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generating client key: %w", err)
	}
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(11),
		Subject:      pkix.Name{CommonName: "untrusted-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, untrustedCACert, &clientKey.PublicKey, untrustedCA)
	if err != nil {
		return nil, nil, fmt.Errorf("creating untrusted client cert: %w", err)
	}
	clientKeyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling client key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER}),
		nil
}

// GenerateServerTLSCert generates a self-signed server TLS certificate for the given domain.
// Returns PEM-encoded certificate and private key.
func GenerateServerTLSCert(t *testing.T, domain string) (certPEM, keyPEM []byte, err error) {
	t.Helper()
	t.Logf("Generating self-signed server TLS cert for domain %s", domain)

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generating server key: %w", err)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject:      pkix.Name{CommonName: domain},
		DNSNames:     []string{domain, "*." + domain},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, serverTemplate, &serverKey.PublicKey, serverKey)
	if err != nil {
		return nil, nil, fmt.Errorf("creating server cert: %w", err)
	}

	serverKeyDER, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling server key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyDER}),
		nil
}

// CreateServerTLSSecret creates a Kubernetes TLS secret in istio-system for the Istio Gateway's credentialName.
// This is needed in non-Gardener environments where no cert-manager provisions the server cert.
func CreateServerTLSSecret(t *testing.T, secretName string, certPEM, keyPEM []byte) error {
	t.Helper()
	namespace := "istio-system"
	t.Logf("Creating server TLS secret %s/%s", namespace, secretName)

	r, err := client.ResourcesClient(t)
	if err != nil {
		return fmt.Errorf("getting resources client: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: namespace},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": certPEM,
			"tls.key": keyPEM,
		},
	}

	if err := r.Create(t.Context(), secret); err != nil {
		return fmt.Errorf("creating server TLS secret: %w", err)
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Deleting server TLS secret %s/%s", namespace, secretName)
		if err := r.Delete(setup.GetCleanupContext(), secret); err != nil && !k8serrors.IsNotFound(err) {
			t.Logf("Failed to delete server TLS secret %s/%s: %v", namespace, secretName, err)
		}
	})

	return nil
}
