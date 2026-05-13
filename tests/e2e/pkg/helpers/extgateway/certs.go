package extgateway

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"
)

// CertBundle holds a CA and a leaf certificate/key pair for mTLS testing.
type CertBundle struct {
	CACertPEM     []byte
	ClientCertPEM []byte
	ClientKeyPEM  []byte
	// Subject is the full DN of the client certificate in Envoy (reversed) format.
	Subject string
}

// GenerateMTLSCerts generates a CA and a leaf client certificate whose subject
// matches RegionSubject so the Lua cert-validation filter accepts it.
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

	// Envoy reverses the ASN.1 RDN order when constructing the peer cert subject string.
	// RegionSubject = "CN=test-client/test-region,L=gateway,OU=test-clients,O=Test Org,C=US"
	// So the ASN.1 order (innermost first) must be C, O, OU, L, CN.
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

	clientKeyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		return CertBundle{}, fmt.Errorf("marshalling client key: %w", err)
	}

	return CertBundle{
		CACertPEM:     pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}),
		ClientCertPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER}),
		ClientKeyPEM:  pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER}),
		Subject:       RegionSubject,
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
