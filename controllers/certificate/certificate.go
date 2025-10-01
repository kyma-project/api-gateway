package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	netutils "k8s.io/utils/net"
)

const (
	keySize = 4096
)

// original code reference: https://github.com/kubernetes/client-go/blob/master/util/cert/cert.go
func GenerateSelfSignedCertificate(host string, alternateIPs []net.IP, alternateDNS []string, maxAge time.Duration) ([]byte, []byte, error) {
	validFrom := time.Now().Add(-time.Hour) // valid an hour earlier to avoid flakes due to clock skew

	// Create CA certificate
	caKey, caCertificate, caDERBytes, err := createCACertificate(host, validFrom, maxAge)
	if err != nil {
		return nil, nil, err
	}

	// Create certificate
	certKey, certDERBytes, err := createCertificate(host, validFrom, maxAge, alternateIPs, alternateDNS, caKey, caCertificate)
	if err != nil {
		return nil, nil, err
	}

	// Certificate followed by the CA certificate
	certBytes, err := encodePEMBlock("CERTIFICATE", certDERBytes, caDERBytes)
	if err != nil {
		return nil, nil, err
	}

	// Key
	keyBytes, err := encodePEMBlock(keyutil.RSAPrivateKeyBlockType, x509.MarshalPKCS1PrivateKey(certKey))
	if err != nil {
		return nil, nil, err
	}

	return certBytes, keyBytes, nil
}

func createCACertificate(host string, validFrom time.Time, maxAge time.Duration) (*rsa.PrivateKey, *x509.Certificate, []byte, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, nil, err
	}

	serial, err := generateRandomSerialNumber()
	if err != nil {
		return nil, nil, nil, err
	}
	caPublicKey := caKey.PublicKey
	publicKeyBytes, err := asn1.Marshal(
		struct {
			N *big.Int
			E int
		}{
			N: caPublicKey.N,
			E: caPublicKey.E,
		})

	if err != nil {
		return nil, nil, nil, err
	}
	sum := sha256.Sum256(publicKeyBytes)
	subjectKeyId := sum[:]

	caTemplate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s-ca@%d", host, time.Now().Unix()),
		},
		NotBefore: validFrom,
		NotAfter:  validFrom.Add(maxAge),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		SubjectKeyId:          subjectKeyId,
	}

	caDERBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caCertificate, err := x509.ParseCertificate(caDERBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	return caKey, caCertificate, caDERBytes, nil
}

func createCertificate(host string, validFrom time.Time, maxAge time.Duration, alternateIPs []net.IP, alternateDNS []string, caKey *rsa.PrivateKey, caCertificate *x509.Certificate) (*rsa.PrivateKey, []byte, error) {
	certKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, nil, err
	}

	serial, err := generateRandomSerialNumber()
	if err != nil {
		return nil, nil, err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s@%d", host, time.Now().Unix()),
		},
		NotBefore: validFrom,
		NotAfter:  validFrom.Add(maxAge),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := netutils.ParseIPSloppy(host); ip != nil {
		certTemplate.IPAddresses = append(certTemplate.IPAddresses, ip)
	} else {
		certTemplate.DNSNames = append(certTemplate.DNSNames, host)
	}

	certTemplate.IPAddresses = append(certTemplate.IPAddresses, alternateIPs...)
	certTemplate.DNSNames = append(certTemplate.DNSNames, alternateDNS...)

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCertificate, &certKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	return certKey, certDERBytes, nil
}

func encodePEMBlock(blockType string, data ...[]byte) ([]byte, error) {
	buffer := bytes.Buffer{}
	for _, d := range data {
		if err := pem.Encode(&buffer, &pem.Block{Type: blockType, Bytes: d}); err != nil {
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

// returns a uniform random value in [0, max-1), then add 1 to serial to make it a uniform random value in [1, max).
func generateRandomSerialNumber() (*big.Int, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64-1))
	if err != nil {
		return nil, err
	}
	return new(big.Int).Add(serial, big.NewInt(1)), nil
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
