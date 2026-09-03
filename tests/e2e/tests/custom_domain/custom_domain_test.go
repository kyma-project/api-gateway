package custom_domain

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"

	e2eclient "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	customdomainhelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/customdomain"
	httpbinhelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpbin"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	oauth2mock "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2/mock"

	"github.com/avast/retry-go/v4"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
)

//go:embed no_auth_apirule.yaml
var APIRuleNoAuth string

//go:embed oauth_apirule.yaml
var APIRuleOAuth string

//go:embed gateway.yaml
var GatewayTemplate string

//go:embed dns_entry.yaml
var DNSEntryTemplate string

//go:embed gcp_credentials_secret.yaml
var GCPCredentialsSecretTemplate string

//go:embed dns_provider.yaml
var DNSProviderTemplate string

//go:embed certificate.yaml
var CertificateTemplate string

const (
	customDomainEnvVar    = "TEST_CUSTOM_DOMAIN"
	gcpSAPathEnvVar       = "TEST_SA_ACCESS_KEY_PATH"
	ingressServiceName    = "istio-ingressgateway"
	ingressServiceNS      = "istio-system"
	dnsResolutionTimeout  = 10 * time.Second
	dnsResolutionAttempts = 100
)

// setupCustomGateway creates a custom Istio Gateway using a Gardener-issued TLS certificate,
// a DNSEntry pointing the wildcard subdomain to the ingress LB (ip/hostname, based on which is stored in the status),
// and returns a "namespace/name" gateway reference for use in APIRule.
// certName is the name of the Gardener Certificate (and its resulting secret) in istio-system.
func setupCustomGateway(t *testing.T, namespace, name, subdomain, certName string) (string, error) {
	t.Helper()
	gatewayName := fmt.Sprintf("custom-domain-%s", name)
	hostWildcard := fmt.Sprintf("*.%s", subdomain)

	_, err := infrahelpers.CreateResourceWithTemplateValues(
		t,
		GatewayTemplate,
		map[string]any{
			"Name":          gatewayName,
			"Host":          hostWildcard,
			"TLSSecretName": certName,
		},
		decoder.MutateNamespace(namespace),
	)
	if err != nil {
		return "", fmt.Errorf("creating custom Istio Gateway %s/%s: %w", namespace, gatewayName, err)
	}

	return fmt.Sprintf("%s/%s", namespace, gatewayName), nil
}

func setupCustomDomainTestBackground(t *testing.T, prefix, oauth2Domain string) (testsetup.TestBackground, error) {
	t.Helper()

	testID, namespace, err := testsetup.CreateNamespaceWithRandomID(
		t,
		testsetup.WithPrefix(prefix),
		testsetup.WithSidecarInjectionEnabled(),
	)
	if err != nil {
		return testsetup.TestBackground{}, err
	}

	svcName, svcPort, err := httpbinhelper.DeployHttpbin(t, namespace)
	if err != nil {
		return testsetup.TestBackground{}, err
	}

	provider, err := oauth2mock.DeployMock(t, namespace, oauth2mock.WithDomain(oauth2Domain))
	if err != nil {
		return testsetup.TestBackground{}, err
	}

	return testsetup.TestBackground{
		TestName:          testID,
		Namespace:         namespace,
		TargetServiceName: svcName,
		TargetServicePort: svcPort,
		Provider:          provider,
	}, nil
}

func TestAPIRuleCustomDomain(t *testing.T) {
	customDomain := os.Getenv(customDomainEnvVar)
	if customDomain == "" {
		t.Fatalf("Failed custom domain tests: %s is not set", customDomainEnvVar)
	}

	gcpSAPath := os.Getenv(gcpSAPathEnvVar)
	if gcpSAPath == "" {
		t.Fatalf("Failed custom domain tests: %s is not set", gcpSAPathEnvVar)
	}

	require.NoError(t, modulehelpers.CreateApiGatewayCR(t), "Failed to create ApiGateway CR")

	gcpSAJson, err := os.ReadFile(gcpSAPath)
	require.NoError(t, err, "Failed to read GCP SA JSON from %s", gcpSAPath)

	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	r, err := e2eclient.ResourcesClient(t)
	require.NoError(t, err, "Failed to create resources client")

	loadBalancerTarget, err := customdomainhelper.GetLoadBalancerTarget(t.Context(), r, ingressServiceName, ingressServiceNS)
	require.NoError(t, err, "Failed to determine ingress load balancer IP")

	t.Logf("Resolved ingress load balancer IP: %s", loadBalancerTarget)

	suiteID := envconf.RandomName("cd", 8)
	gcpSecretName := "gcp-credentials-" + suiteID
	certName := "custom-domain-" + suiteID

	_, err = infrahelpers.CreateResourceWithTemplateValuesSensitive(
		t,
		GCPCredentialsSecretTemplate,
		map[string]any{
			"Name":                 gcpSecretName,
			"EncodedSACredentials": base64.StdEncoding.EncodeToString(gcpSAJson),
		},
		decoder.MutateNamespace("default"),
	)
	require.NoError(t, err, "Failed to create GCP credentials Secret")

	_, err = infrahelpers.CreateResourceWithTemplateValues(
		t,
		DNSProviderTemplate,
		map[string]any{
			"Name":         "custom-domain-" + suiteID,
			"SecretName":   gcpSecretName,
			"ParentDomain": customDomain,
		},
		decoder.MutateNamespace("default"),
	)
	require.NoError(t, err, "Failed to create DNSProvider")

	_, err = infrahelpers.CreateResourceWithTemplateValues(
		t,
		CertificateTemplate,
		map[string]any{
			"Name":      certName,
			"Subdomain": suiteID + "." + customDomain,
		},
	)
	require.NoError(t, err, "Failed to create Gardener Certificate")

	assertCertificateReady(t, r, certName, "istio-system")

	subdomain := fmt.Sprintf("%s.%s", suiteID, customDomain)

	_, err = infrahelpers.CreateResourceWithTemplateValues(
		t,
		DNSEntryTemplate,
		map[string]any{
			"Name":               certName,
			"Subdomain":          subdomain,
			"LoadBalancerTarget": loadBalancerTarget,
		},
		decoder.MutateNamespace("default"),
	)
	require.NoError(t, err, "Failed to create suite-level DNSEntry")

	dnsAttempt, err := customdomainhelper.WaitUntilDNSReady(subdomain, loadBalancerTarget,
		retry.Attempts(dnsResolutionAttempts),
		retry.Delay(dnsResolutionTimeout),
		retry.DelayType(retry.FixedDelay),
	)
	require.NoError(t, err, "Failed waiting for wildcard DNS to resolve")
	t.Logf("DNS resolved successfully on attempt #%d", dnsAttempt)

	t.Run("Calling an unsecured API endpoint with custom domain", func(t *testing.T) {

		testBackground, err := setupCustomDomainTestBackground(t, "custom-domain-noauth", kymaGatewayDomain)
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		host := fmt.Sprintf("httpbin-%s.%s", testBackground.TestName, subdomain)

		gatewayRef, err := setupCustomGateway(t, testBackground.Namespace, testBackground.TestName, subdomain, certName)
		require.NoError(t, err, "Failed to create custom Istio Gateway")

		_, err = infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuth,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        host,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     gatewayRef,
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s/headers", host)

		ipfamily.ForEachDialNetwork(t, "custom-domain-noauth", nil, func(t *testing.T, network string, _ *http.Client) {
			statusCode, _, _, err := testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithoutToken(), oauth2.WithNetwork(network))
			require.NoError(t, err, "Request without token should succeed for noAuth endpoint")
			assert.GreaterOrEqual(t, statusCode, http.StatusOK)
			assert.Less(t, statusCode, http.StatusMultipleChoices)

			statusCode, _, _, err = testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithTokenOverride("any-token"), oauth2.WithNetwork(network))
			require.NoError(t, err, "Request with arbitrary token should succeed for noAuth endpoint")
			assert.GreaterOrEqual(t, statusCode, http.StatusOK)
			assert.Less(t, statusCode, http.StatusMultipleChoices)
		})
	})

	t.Run("Calling a secured API with JWT and custom domain", func(t *testing.T) {
		testBackground, err := setupCustomDomainTestBackground(t, "custom-domain-jwt", kymaGatewayDomain)
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		host := fmt.Sprintf("httpbin-%s.%s", testBackground.TestName, subdomain)

		gatewayRef, err := setupCustomGateway(t, testBackground.Namespace, testBackground.TestName, subdomain, certName)
		require.NoError(t, err, "Failed to create custom Istio Gateway")

		_, err = infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleOAuth,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        host,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     gatewayRef,
				"Issuer":      testBackground.Provider.GetIssuerURL(),
				"JwksUri":     testBackground.Provider.GetJwksURI(),
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s/headers", host)

		ipfamily.ForEachDialNetwork(t, "custom-domain-jwt", nil, func(t *testing.T, network string, _ *http.Client) {
			statusCode, _, _, err := testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithoutToken(), oauth2.WithNetwork(network))
			require.NoError(t, err, "Request without token should be rejected for JWT-secured endpoint")
			assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, statusCode, "Expected 401 or 403 for missing token")

			statusCode, _, _, err = testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithTokenOverride("any-token"), oauth2.WithNetwork(network))
			require.NoError(t, err, "Request with invalid token should be rejected for JWT-secured endpoint")
			assert.Contains(t, []int{http.StatusUnauthorized, http.StatusForbidden}, statusCode, "Expected 401 or 403 for invalid token")

			validToken, err := testBackground.Provider.GetToken(t)
			require.NoError(t, err, "Fetching valid token should succeed")

			statusCode, _, _, err = testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithTokenOverride(validToken), oauth2.WithNetwork(network))
			require.NoError(t, err, "Request with valid JWT token should succeed")
			assert.GreaterOrEqual(t, statusCode, http.StatusOK)
			assert.Less(t, statusCode, http.StatusMultipleChoices)
		})
	})
}

func assertCertificateReady(t *testing.T, r *resources.Resources, name, namespace string) {
	t.Helper()
	cert := &unstructured.Unstructured{}
	cert.SetGroupVersionKind(schema.GroupVersionKind{Group: "cert.gardener.cloud", Version: "v1alpha1", Kind: "Certificate"})
	cert.SetName(name)
	cert.SetNamespace(namespace)
	require.NoError(t,
		wait.For(
			conditions.New(r).ResourceMatch(cert, func(obj k8s.Object) bool {
				u, ok := obj.(*unstructured.Unstructured)
				if !ok {
					return false
				}
				state, _, _ := unstructured.NestedString(u.Object, "status", "state")
				return state == "Ready"
			}),
			wait.WithTimeout(15*time.Minute),
		),
		"Gardener Certificate %s/%s did not become ready", namespace, name,
	)
}
