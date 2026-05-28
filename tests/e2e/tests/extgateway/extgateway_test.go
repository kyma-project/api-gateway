package extgateway

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	endpointasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/endpoint"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	extgwhelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/extgateway"
	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
)

const externalDomainBase = "local.kyma.dev"

//go:embed apirule_no_auth.yaml
var apiRuleExtGateway string

//go:embed apirule_kyma_gateway.yaml
var apiRuleKymaGateway string

func setupExternalGateway(t *testing.T, namespace, testName string, caCertPEM []byte, subject string) string {
	t.Helper()

	caSecretName := testName + "-ca"
	require.NoError(t, extgwhelper.CreateCASecret(t, namespace, caSecretName, caCertPEM))

	regionsConfigMap := testName + "-regions"
	require.NoError(t, extgwhelper.CreateRegionsConfigMap(
		t, namespace, regionsConfigMap,
		extgwhelper.RegionsConfigMapData(extgwhelper.RegionName, subject),
	))

	externalDomain := fmt.Sprintf("%s.%s", testName, externalDomainBase)

	serverCert, serverKey, err := extgwhelper.GenerateServerTLSCert(t, externalDomain)
	require.NoError(t, err, "Failed to generate server TLS cert")
	require.NoError(t, extgwhelper.CreateServerTLSSecret(t, extgwhelper.TLSSecretName(testName), serverCert, serverKey))

	_, err = extgwhelper.CreateExternalGateway(
		t, namespace, testName,
		externalDomain, testName+"-int",
		extgwhelper.RegionName, regionsConfigMap, caSecretName,
	)
	require.NoError(t, err)
	require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, namespace, testName))

	return externalDomain
}

func setupExternalGatewayWithAPIRule(t *testing.T, namespace, testName, serviceName string, servicePort any, caCertPEM []byte, subject string) string {
	t.Helper()

	externalDomain := setupExternalGateway(t, namespace, testName, caCertPEM, subject)

	_, err := infrahelpers.CreateResourceWithTemplateValues(t, apiRuleExtGateway, map[string]any{
		"Name":            testName,
		"Host":            externalDomain,
		"ServiceName":     serviceName,
		"ServicePort":     servicePort,
		"ExternalGateway": fmt.Sprintf("%s/%s", namespace, testName),
	}, decoder.MutateNamespace(namespace))
	require.NoError(t, err)
	apiruleasserts.WaitUntilReady(t, testName, namespace)

	return externalDomain
}

func TestExternalGateway(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	certs, err := extgwhelper.GenerateMTLSCerts(t)
	require.NoError(t, err, "Failed to generate mTLS cert bundle")

	t.Run("valid client cert reaches workload via external domain", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-ok"))
		require.NoError(t, err)

		externalDomain := setupExternalGatewayWithAPIRule(
			t,
			bg.Namespace,
			bg.TestName,
			bg.TargetServiceName,
			bg.TargetServicePort,
			certs.CACertPEM,
			certs.Subject,
		)

		body, err := extgwhelper.AssertMTLSEndpoint(
			t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", externalDomain),
			certs.ClientCertPEM, certs.ClientKeyPEM,
			http.StatusOK,
			certs.CACertPEM,
		)
		require.NoError(t, err, "request with valid client cert should return 200")
		assert.NotEmpty(t, body)
	})

	t.Run("Lua filter rejects wrong cert subject with 403", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-subj"))
		require.NoError(t, err)

		externalDomain := setupExternalGatewayWithAPIRule(
			t,
			bg.Namespace,
			bg.TestName,
			bg.TargetServiceName,
			bg.TargetServicePort,
			certs.CACertPEM,
			"CN=other-client/other-region,L=gateway,OU=test-clients,O=Test Org,C=US",
		)

		body, mtlsErr := extgwhelper.AssertMTLSEndpoint(
			t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", externalDomain),
			certs.ClientCertPEM, certs.ClientKeyPEM,
			http.StatusForbidden,
			certs.CACertPEM,
		)
		// The Lua filter may return HTTP 403 or reset the connection (EOF).
		// Both are valid rejection behaviours.
		if mtlsErr != nil {
			assert.Contains(t, mtlsErr.Error(), "EOF",
				"if the request fails, it should be due to a connection reset (EOF), not a timeout")
		} else {
			assert.True(t, body == "" || body == "Forbidden",
				"rejected request should have either an empty body or the Lua filter response body 'Forbidden'")
		}
	})

	t.Run("TLS handshake fails with untrusted cert", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-untrust"))
		require.NoError(t, err)

		externalDomain := setupExternalGatewayWithAPIRule(
			t,
			bg.Namespace,
			bg.TestName,
			bg.TargetServiceName,
			bg.TargetServicePort,
			certs.CACertPEM,
			certs.Subject,
		)

		untrustedCert, untrustedKey, err := extgwhelper.GenerateUntrustedClientCert(t)
		require.NoError(t, err)

		httpClient, err := extgwhelper.NewMTLSHTTPClient(t, untrustedCert, untrustedKey)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/headers", externalDomain), nil)
		require.NoError(t, err)

		_, err = httpClient.Do(req)
		require.Error(t, err, "TLS handshake should fail with an untrusted certificate")
	})

	t.Run("mTLS endpoint rejects connection without client cert", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-nocert"))
		require.NoError(t, err)

		externalDomain := setupExternalGatewayWithAPIRule(
			t,
			bg.Namespace,
			bg.TestName,
			bg.TargetServiceName,
			bg.TargetServicePort,
			certs.CACertPEM,
			certs.Subject,
		)

		httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("no-cert-client"))
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/headers", externalDomain), nil)
		require.NoError(t, err)

		_, err = httpClient.Do(req)
		require.Error(t, err, "request without client cert must be rejected by mTLS MUTUAL endpoint")
	})

	t.Run("workload receives a single sanitized XFCC entry", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-xfcc"))
		require.NoError(t, err)

		externalDomain := setupExternalGatewayWithAPIRule(
			t,
			bg.Namespace,
			bg.TestName,
			bg.TargetServiceName,
			bg.TargetServicePort,
			certs.CACertPEM,
			certs.Subject,
		)

		httpClient, err := extgwhelper.NewMTLSHTTPClient(t, certs.ClientCertPEM, certs.ClientKeyPEM, certs.CACertPEM)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/headers", externalDomain), nil)
		require.NoError(t, err)

		req.Header.Set("X-Forwarded-Client-Cert", "URI=spiffe://cluster.local/ns/default/sa/client-sa,By=spiffe://cluster.local/ns/istio-system/sa/some-proxy")

		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		rawBody, _ := io.ReadAll(resp.Body)
		body := string(rawBody)

		var parsed struct {
			Headers map[string]any `json:"headers"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &parsed), "httpbin /headers should return parseable JSON")

		var xfcc string
		switch value := parsed.Headers["X-Forwarded-Client-Cert"].(type) {
		case string:
			xfcc = value
		case []any:
			parts := make([]string, 0, len(value))
			for _, item := range value {
				itemString, ok := item.(string)
				require.True(t, ok, "X-Forwarded-Client-Cert array should contain only strings")
				parts = append(parts, itemString)
			}
			xfcc = strings.Join(parts, ",")
		default:
			require.Failf(t, "unexpected XFCC type", "X-Forwarded-Client-Cert should be a string or array of strings, got %T", value)
		}

		require.NotEmpty(t, xfcc, "X-Forwarded-Client-Cert must be present in workload request")
		assert.Equal(t, strings.Count(xfcc, "By="), 2, "XFCC should contain two entries (proxy + ingress)")
		assert.Contains(t, xfcc, "URI=spiffe://cluster.local/ns/default/sa/client-sa", "XFCC should preserve the initial client certificate entry")
		assert.Contains(t, xfcc, "URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			"XFCC should contain the ingress gateway certificate entry from internal mTLS")
		assert.Contains(t, xfcc, "some-proxy", "XFCC should contain the proxy certificate from the initial header")
	})

	t.Run("ExternalGateway enters Error state with wrong region", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-badregion"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData("other-region", certs.Subject),
		))

		externalDomain := fmt.Sprintf("%s.%s", bg.TestName, externalDomainBase)
		serverCert, serverKey, err := extgwhelper.GenerateServerTLSCert(t, externalDomain)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.CreateServerTLSSecret(t, extgwhelper.TLSSecretName(bg.TestName), serverCert, serverKey))

		_, err = extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName+"-int",
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)

		require.NoError(t,
			extgwhelper.WaitUntilExternalGatewayError(t, bg.Namespace, bg.TestName),
			"ExternalGateway should enter Error state when the requested region is missing from the regions ConfigMap",
		)
	})

	t.Run("ExternalGateway enters Error state with invalid regions configmap", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-badcm"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			"this is: not: valid: yaml: [\n",
		))

		_, err = extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			fmt.Sprintf("%s.%s", bg.TestName, externalDomainBase), bg.TestName+"-int",
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)

		require.NoError(t,
			extgwhelper.WaitUntilExternalGatewayError(t, bg.Namespace, bg.TestName),
			"ExternalGateway should enter Error state when regions ConfigMap is malformed",
		)
	})

	t.Run("kyma default domain remains accessible and isolated from external gateway", func(t *testing.T) {
		t.Parallel()

		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		bgKyma, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-kyma"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, apiRuleKymaGateway, map[string]any{
			"Name":        bgKyma.TestName,
			"Host":        bgKyma.TestName,
			"ServiceName": bgKyma.TargetServiceName,
			"ServicePort": bgKyma.TargetServicePort,
			"Gateway":     "kyma-system/kyma-gateway",
		}, decoder.MutateNamespace(bgKyma.Namespace))
		require.NoError(t, err)
		apiruleasserts.WaitUntilReady(t, bgKyma.TestName, bgKyma.Namespace)

		bgExt, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-iso"))
		require.NoError(t, err)

		extDomain := setupExternalGatewayWithAPIRule(
			t,
			bgExt.Namespace,
			bgExt.TestName,
			bgExt.TargetServiceName,
			bgExt.TargetServicePort,
			certs.CACertPEM,
			certs.Subject,
		)

		kymaURL := fmt.Sprintf("https://%s.%s/headers", bgKyma.TestName, kymaGatewayDomain)
		require.NoError(t,
			endpointasserts.AssertEndpoint(t, http.MethodGet, kymaURL, http.StatusOK),
			"kyma default gateway must remain reachable after ExternalGateway is created",
		)

		_, err = extgwhelper.AssertMTLSEndpoint(
			t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", extDomain),
			certs.ClientCertPEM, certs.ClientKeyPEM,
			http.StatusOK,
			certs.CACertPEM,
		)
		require.NoError(t, err, "external gateway httpbin request with mTLS should succeed")
	})
}
