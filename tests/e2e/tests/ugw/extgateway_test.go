package extgateway

import (
	_ "embed"
	"encoding/json"
	"fmt"
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
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
)

//go:embed apirule_no_auth.yaml
var apiRuleExtGateway string

//go:embed apirule_kyma_gateway.yaml
var apiRuleKymaGateway string

// TestExternalGateway runs the full external gateway e2e suite.
//
// The test process acts as the external client: it connects directly to the cluster's
// Istio ingress and presents a client TLS certificate.  No in-cluster curl pods are
// used — all HTTP calls are made from the test process itself.
//
// Cluster prerequisites:
//   - APIGateway operator and ExternalGateway controller deployed
//   - Istio installed
//   - The external domain resolves to the Istio ingress (or /etc/hosts is configured)
func TestExternalGateway(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	// Generate CA + client cert whose subject matches the region entry.
	// The CA is stored in the cluster so Istio validates the mTLS handshake.
	// The client cert is used by the test process when making HTTPS calls.
	certs, err := extgwhelper.GenerateMTLSCerts(t)
	require.NoError(t, err, "Failed to generate mTLS cert bundle")

	// -------------------------------------------------------------------------
	// Happy path: valid client cert + correct region → workload responds 200
	// -------------------------------------------------------------------------
	t.Run("happy path: valid client cert reaches workload via external domain", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-ok"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName, certs.Subject),
		))

		externalDomain := fmt.Sprintf("%s.ext.example.com", bg.TestName)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName,
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, bg.Namespace, bg.TestName))

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, apiRuleExtGateway, map[string]any{
			"Name":            bg.TestName,
			"Host":            externalDomain,
			"ServiceName":     bg.TargetServiceName,
			"ServicePort":     bg.TargetServicePort,
			"ExternalGateway": extgwhelper.ExternalGatewayRef(bg.Namespace, eg),
		}, decoder.MutateNamespace(bg.Namespace))
		require.NoError(t, err)
		apiruleasserts.WaitUntilReady(t, bg.TestName, bg.Namespace)

		body, err := extgwhelper.AssertMTLSEndpoint(
			t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", externalDomain),
			certs.ClientCertPEM, certs.ClientKeyPEM,
			http.StatusOK,
		)
		require.NoError(t, err, "request with valid client cert should return 200")
		assert.NotEmpty(t, body)
	})

	// -------------------------------------------------------------------------
	// Wrong cert subject: cert is trusted by CA but subject not in region list
	// -------------------------------------------------------------------------
	t.Run("wrong cert subject: Lua filter rejects with 403", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-subj"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		// ConfigMap references a *different* subject — our cert's subject won't match.
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName,
				"CN=other-client/other-region,L=gateway,OU=test-clients,O=Test Org,C=US"),
		))

		externalDomain := fmt.Sprintf("%s.ext.example.com", bg.TestName)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName,
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, bg.Namespace, bg.TestName))

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, apiRuleExtGateway, map[string]any{
			"Name":            bg.TestName,
			"Host":            externalDomain,
			"ServiceName":     bg.TargetServiceName,
			"ServicePort":     bg.TargetServicePort,
			"ExternalGateway": extgwhelper.ExternalGatewayRef(bg.Namespace, eg),
		}, decoder.MutateNamespace(bg.Namespace))
		require.NoError(t, err)
		apiruleasserts.WaitUntilReady(t, bg.TestName, bg.Namespace)

		_, err = extgwhelper.AssertMTLSEndpoint(
			t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", externalDomain),
			certs.ClientCertPEM, certs.ClientKeyPEM,
			http.StatusForbidden,
		)
		require.NoError(t, err, "mismatched cert subject should result in HTTP 403 from Lua filter")
	})

	// -------------------------------------------------------------------------
	// Untrusted certificate: signed by a different CA → TLS handshake fails
	// -------------------------------------------------------------------------
	t.Run("untrusted cert: TLS handshake fails", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-untrust"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName, certs.Subject),
		))

		externalDomain := fmt.Sprintf("%s.ext.example.com", bg.TestName)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName,
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, bg.Namespace, bg.TestName))

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, apiRuleExtGateway, map[string]any{
			"Name":            bg.TestName,
			"Host":            externalDomain,
			"ServiceName":     bg.TargetServiceName,
			"ServicePort":     bg.TargetServicePort,
			"ExternalGateway": extgwhelper.ExternalGatewayRef(bg.Namespace, eg),
		}, decoder.MutateNamespace(bg.Namespace))
		require.NoError(t, err)
		apiruleasserts.WaitUntilReady(t, bg.TestName, bg.Namespace)

		untrustedCert, untrustedKey, err := extgwhelper.GenerateUntrustedClientCert(t)
		require.NoError(t, err)

		httpClient, err := extgwhelper.NewMTLSHTTPClient(t, untrustedCert, untrustedKey)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/headers", externalDomain), nil)
		require.NoError(t, err)

		_, err = httpClient.Do(req)
		require.Error(t, err, "TLS handshake should fail with an untrusted certificate")
	})

	// -------------------------------------------------------------------------
	// No client certificate: mTLS MUTUAL mode rejects the connection
	// -------------------------------------------------------------------------
	t.Run("no client cert: mTLS endpoint rejects connection", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-nocert"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName, certs.Subject),
		))

		externalDomain := fmt.Sprintf("%s.ext.example.com", bg.TestName)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName,
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, bg.Namespace, bg.TestName))

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, apiRuleExtGateway, map[string]any{
			"Name":            bg.TestName,
			"Host":            externalDomain,
			"ServiceName":     bg.TargetServiceName,
			"ServicePort":     bg.TargetServicePort,
			"ExternalGateway": extgwhelper.ExternalGatewayRef(bg.Namespace, eg),
		}, decoder.MutateNamespace(bg.Namespace))
		require.NoError(t, err)
		apiruleasserts.WaitUntilReady(t, bg.TestName, bg.Namespace)

		// Plain HTTPS client — no client cert.  Istio mTLS MUTUAL must reject.
		err = endpointasserts.AssertEndpoint(t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", externalDomain),
			http.StatusOK,
		)
		require.Error(t, err, "request without client cert must be rejected by mTLS MUTUAL endpoint")
	})

	// -------------------------------------------------------------------------
	// XFCC forward-only: client cert forwarded, ingress cert stripped
	// -------------------------------------------------------------------------
	t.Run("XFCC forward-only: client cert forwarded to workload, ingress cert absent", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-xfcc"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName, certs.Subject),
		))

		externalDomain := fmt.Sprintf("%s.ext.example.com", bg.TestName)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName,
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, bg.Namespace, bg.TestName))

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, apiRuleExtGateway, map[string]any{
			"Name":            bg.TestName,
			"Host":            externalDomain,
			"ServiceName":     bg.TargetServiceName,
			"ServicePort":     bg.TargetServicePort,
			"ExternalGateway": extgwhelper.ExternalGatewayRef(bg.Namespace, eg),
		}, decoder.MutateNamespace(bg.Namespace))
		require.NoError(t, err)
		apiruleasserts.WaitUntilReady(t, bg.TestName, bg.Namespace)

		// httpbin /headers echoes all incoming request headers in a JSON body.
		body, err := extgwhelper.AssertMTLSEndpoint(
			t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", externalDomain),
			certs.ClientCertPEM, certs.ClientKeyPEM,
			http.StatusOK,
		)
		require.NoError(t, err)

		// Parse JSON response {"headers": {"X-Forwarded-Client-Cert": "..."}}
		var parsed struct {
			Headers map[string]string `json:"headers"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &parsed), "httpbin /headers should return parseable JSON")

		xfcc := parsed.Headers["X-Forwarded-Client-Cert"]
		// XFCC must be present (FORWARD_ONLY mode forwards the downstream client cert).
		require.NotEmpty(t, xfcc, "X-Forwarded-Client-Cert must be present in workload request")

		// The forwarded XFCC must contain the *client* cert CN, not the Istio ingress cert.
		assert.Contains(t, xfcc, "test-client", "XFCC should carry the client cert CN")

		// FORWARD_ONLY strips the ingressgateway's own cert from the header; there should
		// be only one entry (no stacked XFCC values from the ingress itself).
		assert.Equal(t, 1, strings.Count(xfcc, "By="), "XFCC should contain exactly one cert entry (client cert, no ingress cert appended)")
	})

	// -------------------------------------------------------------------------
	// Invalid regions ConfigMap: ExternalGateway enters Error state
	// -------------------------------------------------------------------------
	t.Run("invalid regions configmap: ExternalGateway enters Error state", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-badcm"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			// intentionally broken YAML
			"this is: not: valid: yaml: [\n",
		))

		_, err = extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			fmt.Sprintf("%s.ext.example.com", bg.TestName), bg.TestName,
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)

		require.NoError(t,
			extgwhelper.WaitUntilExternalGatewayError(t, bg.Namespace, bg.TestName),
			"ExternalGateway should enter Error state when regions ConfigMap is malformed",
		)
	})

	// -------------------------------------------------------------------------
	// Kyma default domain: external gateway must not affect the Kyma gateway
	// -------------------------------------------------------------------------
	t.Run("kyma default domain remains accessible and isolated from external gateway", func(t *testing.T) {
		t.Parallel()

		// Expose a workload via the default Kyma gateway (no client cert required).
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

		// Separately create an ExternalGateway in the cluster to verify it doesn't
		// pollute the Kyma default gateway routing.
		bgExt, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-iso"))
		require.NoError(t, err)

		caSecretName := bgExt.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bgExt.Namespace, caSecretName, certs.CACertPEM))
		regionsConfigMap := bgExt.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bgExt.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName, certs.Subject),
		))
		_, err = extgwhelper.CreateExternalGateway(
			t, bgExt.Namespace, bgExt.TestName,
			fmt.Sprintf("%s.ext.example.com", bgExt.TestName), bgExt.TestName,
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, bgExt.Namespace, bgExt.TestName))

		// The Kyma-domain workload must still respond with 200 without a client cert.
		kymaURL := fmt.Sprintf("https://%s.%s/headers", bgKyma.TestName, kymaGatewayDomain)
		require.NoError(t,
			endpointasserts.AssertEndpoint(t, http.MethodGet, kymaURL, http.StatusOK),
			"kyma default gateway must remain reachable after ExternalGateway is created",
		)
	})
}
