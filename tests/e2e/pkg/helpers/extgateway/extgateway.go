package extgateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RegionName    = "test-region"
	RegionSubject = "CN=test-client/test-region,L=gateway,OU=test-clients,O=Test Org,C=US"
)

// RegionsConfigMapData returns the YAML content for a regions ConfigMap.
func RegionsConfigMapData(region, subject string) string {
	return fmt.Sprintf(`regions:
  - name: "%s"
    ips:
      - 10.0.0.1
    subjects:
      - "%s"
`, region, subject)
}

// CreateCASecret creates a Kubernetes Secret containing the CA certificate.
func CreateCASecret(t *testing.T, namespace, name string, caCertPEM []byte) error {
	t.Helper()
	t.Logf("Creating CA secret %s/%s", namespace, name)

	r, err := client.ResourcesClient(t)
	if err != nil {
		return fmt.Errorf("getting resources client: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Data:       map[string][]byte{"ca.crt": caCertPEM},
	}

	if err := r.Create(t.Context(), secret); err != nil {
		return fmt.Errorf("creating CA secret: %w", err)
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Deleting CA secret %s/%s", namespace, name)
		if err := r.Delete(setup.GetCleanupContext(), secret); err != nil && !k8serrors.IsNotFound(err) {
			t.Logf("Failed to delete CA secret %s/%s: %v", namespace, name, err)
		}
	})

	return nil
}

// CreateRegionsConfigMap creates the ConfigMap with region metadata.
func CreateRegionsConfigMap(t *testing.T, namespace, name, data string) error {
	t.Helper()
	t.Logf("Creating regions ConfigMap %s/%s", namespace, name)

	r, err := client.ResourcesClient(t)
	if err != nil {
		return fmt.Errorf("getting resources client: %w", err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Data:       map[string]string{"regions.yaml": data},
	}

	if err := r.Create(t.Context(), cm); err != nil {
		return fmt.Errorf("creating regions ConfigMap: %w", err)
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Deleting regions ConfigMap %s/%s", namespace, name)
		if err := r.Delete(setup.GetCleanupContext(), cm); err != nil && !k8serrors.IsNotFound(err) {
			t.Logf("Failed to delete regions ConfigMap %s/%s: %v", namespace, name, err)
		}
	})

	return nil
}

// CreateExternalGateway creates an ExternalGateway CR.
func CreateExternalGateway(t *testing.T, namespace, name, externalDomain, kymaSubdomain, region, regionsConfigMap, caSecretName string) (*externalv1alpha1.ExternalGateway, error) {
	t.Helper()
	t.Logf("Creating ExternalGateway %s/%s", namespace, name)

	r, err := client.ResourcesClient(t)
	if err != nil {
		return nil, fmt.Errorf("getting resources client: %w", err)
	}

	if err := externalv1alpha1.AddToScheme(r.GetScheme()); err != nil {
		return nil, fmt.Errorf("adding ExternalGateway scheme: %w", err)
	}

	eg := &externalv1alpha1.ExternalGateway{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: externalv1alpha1.ExternalGatewaySpec{
			ExternalDomain: externalDomain,
			InternalDomain: externalv1alpha1.InternalDomainConfig{KymaSubdomain: kymaSubdomain},
			Region:         region,
			RegionsConfigMap: regionsConfigMap,
			CASecretRef: &corev1.SecretReference{
				Name:      caSecretName,
				Namespace: namespace,
			},
		},
	}

	if err := r.Create(t.Context(), eg); err != nil {
		return nil, fmt.Errorf("creating ExternalGateway: %w", err)
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Deleting ExternalGateway %s/%s", namespace, name)
		if err := r.Delete(setup.GetCleanupContext(), eg); err != nil && !k8serrors.IsNotFound(err) {
			t.Logf("Failed to delete ExternalGateway %s/%s: %v", namespace, name, err)
		}
	})

	return eg, nil
}

// WaitUntilExternalGatewayReady polls until the ExternalGateway reaches Ready state.
func WaitUntilExternalGatewayReady(t *testing.T, namespace, name string) error {
	t.Helper()
	t.Logf("Waiting for ExternalGateway %s/%s to be Ready", namespace, name)

	r, err := client.ResourcesClient(t)
	if err != nil {
		return fmt.Errorf("getting resources client: %w", err)
	}

	eg := &externalv1alpha1.ExternalGateway{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}

	return wait.For(
		conditions.New(r).ResourceMatch(eg, func(obj k8s.Object) bool {
			egObj, ok := obj.(*externalv1alpha1.ExternalGateway)
			if !ok {
				return false
			}
			return egObj.Status.State == externalv1alpha1.Ready
		}),
		wait.WithTimeout(2*time.Minute),
	)
}

// WaitUntilExternalGatewayError polls until the ExternalGateway reaches Error state.
func WaitUntilExternalGatewayError(t *testing.T, namespace, name string) error {
	t.Helper()
	t.Logf("Waiting for ExternalGateway %s/%s to be in Error state", namespace, name)

	r, err := client.ResourcesClient(t)
	if err != nil {
		return fmt.Errorf("getting resources client: %w", err)
	}

	eg := &externalv1alpha1.ExternalGateway{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}

	return wait.For(
		conditions.New(r).ResourceMatch(eg, func(obj k8s.Object) bool {
			egObj, ok := obj.(*externalv1alpha1.ExternalGateway)
			if !ok {
				return false
			}
			return egObj.Status.State == externalv1alpha1.Error
		}),
		wait.WithTimeout(2*time.Minute),
	)
}

// GetExternalGateway fetches the current ExternalGateway state.
func GetExternalGateway(t *testing.T, namespace, name string) (*externalv1alpha1.ExternalGateway, error) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	if err != nil {
		return nil, fmt.Errorf("getting resources client: %w", err)
	}

	eg := &externalv1alpha1.ExternalGateway{}
	if err := r.Get(t.Context(), name, namespace, eg); err != nil {
		return nil, fmt.Errorf("getting ExternalGateway %s/%s: %w", namespace, name, err)
	}

	return eg, nil
}

// ExternalGatewayRef returns the "namespace/name" reference for use in an APIRule externalGateway field.
func ExternalGatewayRef(namespace string, eg *externalv1alpha1.ExternalGateway) string {
	return fmt.Sprintf("%s/%s", namespace, eg.Name)
}

// NewMTLSHTTPClient returns an *http.Client that presents the given TLS certificate pair
// on every request and skips server certificate verification (matching the existing
// test helper pattern — clusters use self-signed certs).
func NewMTLSHTTPClient(t *testing.T, certPEM, keyPEM []byte) (*http.Client, error) {
	t.Helper()

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("loading key pair: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true, //nolint:gosec // test clusters use self-signed certs
		},
	}

	return &http.Client{
		Transport: httphelper.TestLogTransportWrapper(t, "mtls-client", transport),
		Timeout:   15 * time.Second,
	}, nil
}

// AssertMTLSEndpoint makes an HTTP request with a client certificate and asserts the
// expected status code.  Returns the full response body and any transport-level error.
func AssertMTLSEndpoint(t *testing.T, method, url string, certPEM, keyPEM []byte, expectedCode int) (body string, err error) {
	t.Helper()

	httpClient, err := NewMTLSHTTPClient(t, certPEM, keyPEM)
	if err != nil {
		return "", fmt.Errorf("building mTLS HTTP client: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("performing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	rawBody, _ := io.ReadAll(resp.Body)
	bodyStr := string(rawBody)

	if resp.StatusCode != expectedCode {
		return bodyStr, fmt.Errorf("expected HTTP %d, got %d", expectedCode, resp.StatusCode)
	}

	return bodyStr, nil
}

// WaitUntilSecretExists polls until the named Secret appears in the given namespace.
func WaitUntilSecretExists(t *testing.T, namespace, name string) error {
	t.Helper()
	t.Logf("Waiting for Secret %s/%s", namespace, name)

	r, err := client.ResourcesClient(t)
	if err != nil {
		return fmt.Errorf("getting resources client: %w", err)
	}

	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		secret := &corev1.Secret{}
		if err := r.Get(context.Background(), name, namespace, secret); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("secret %s/%s did not appear within timeout", namespace, name)
}
