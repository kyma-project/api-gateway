package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/cert"
)

const (
	keySize = 4096
)

// original code reference: https://github.com/kubernetes/client-go/blob/master/util/cert/cert.go
func GenerateSelfSignedCertificate(host string, alternateIPs []net.IP, alternateDNS []string, maxAge time.Duration) ([]byte, []byte, error) {
	// Generate CA key pair
	caKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	// Generate CA certificate
	caCertificate, err := generateCertificate(host+"-ca", time.Now(), maxAge, caKey, true, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	// Generate server key pair
	serverKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	// Generate server certificate
	serverCertificate, err := generateCertificate(host, time.Now(), maxAge, serverKey, false, alternateIPs, alternateDNS)
	if err != nil {
		return nil, nil, err
	}

	// Encode certificates to PEM format
	certPEM := encodeToPEM(serverCertificate, caCertificate)
	keyPEM := encodeToPEM(x509.MarshalPKCS1PrivateKey(serverKey), nil)

	return certPEM, keyPEM, nil
}

func generateCertificate(commonName string, validFrom time.Time, maxAge time.Duration, key *rsa.PrivateKey, isCA bool, alternateIPs []net.IP, alternateDNS []string) ([]byte, error) {
	serial, err := generateRandomSerialNumber()
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             validFrom,
		NotAfter:              validFrom.Add(maxAge),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
		template.IsCA = true
	}

	if len(alternateIPs) > 0 {
		template.IPAddresses = append(template.IPAddresses, alternateIPs...)
	} else {
		template.DNSNames = append(template.DNSNames, alternateDNS...)
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
}

func encodeToPEM(certs ...[]byte) []byte {
	var pemData []byte
	for _, cert := range certs {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert,
		}
		pemData = append(pemData, pem.EncodeToMemory(block)...)
	}
	return pemData
}

func verifySecret(s *corev1.Secret) error {
	if !hasRequiredKeys(s.Data, []string{certificateName, keyName}) {
		return fmt.Errorf("secret does not have required keys: %s, %s", certificateName, keyName)
	}

	if err := verifyCertificate(s.Data[certificateName]); err != nil {
		return err
	}

	if err := verifyKey(s.Data[keyName]); err != nil {
		return err
	}

	return nil
}

func verifyCertificate(c []byte) error {
	certificate, err := cert.ParseCertsPEM(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse certificate data")
	}

	// certificate is self signed, we use it as a root cert
	root, err := cert.NewPoolFromBytes(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse root certificate data")
	}

	// make sure the certificate is valid for predefined duration
	_, err = certificate[0].Verify(x509.VerifyOptions{
		CurrentTime: time.Now().Add(untilRenewal),
		Roots:       root,
	})

	if err != nil {
		return errors.Wrap(err, "certificate verification failed")
	}

	return nil
}

func verifyKey(k []byte) error {
	b, _ := pem.Decode(k)
	key, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse key data")
	}

	if err = key.Validate(); err != nil {
		return errors.Wrap(err, "key verification failed")
	}

	return nil
}

func hasRequiredKeys(data map[string][]byte, keys []string) bool {
	if data == nil {
		return false
	}

	for _, key := range keys {
		if _, ok := data[key]; !ok {
			return false
		}
	}

	return true
}

// returns a uniform random value in [0, max-1), then add 1 to serial to make it a uniform random value in [1, max).
func generateRandomSerialNumber() (*big.Int, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64-1))
	if err != nil {
		return nil, err
	}
	return new(big.Int).Add(serial, big.NewInt(1)), nil
}
