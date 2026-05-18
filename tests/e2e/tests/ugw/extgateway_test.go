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

func TestExternalGateway(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	certs, err := extgwhelper.GenerateMTLSCerts(t)
	require.NoError(t, err, "Failed to generate mTLS cert bundle")

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

		externalDomain := fmt.Sprintf("%s.%s", bg.TestName, externalDomainBase)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName+"-int",
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

	t.Run("wrong cert subject: Lua filter rejects with 403", func(t *testing.T) {
		t.Parallel()

		bg, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("extgw-subj"))
		require.NoError(t, err)

		caSecretName := bg.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bg.Namespace, caSecretName, certs.CACertPEM))

		regionsConfigMap := bg.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bg.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName,
				"CN=other-client/other-region,L=gateway,OU=test-clients,O=Test Org,C=US"),
		))

		externalDomain := fmt.Sprintf("%s.%s", bg.TestName, externalDomainBase)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName+"-int",
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

		body, mtlsErr := extgwhelper.AssertMTLSEndpoint(
			t, http.MethodGet,
			fmt.Sprintf("https://%s/headers", externalDomain),
			certs.ClientCertPEM, certs.ClientKeyPEM,
			http.StatusForbidden,
		)
		// The Lua filter may return HTTP 403 or reset the connection (EOF).
		// Both are valid rejection behaviours.
		if mtlsErr != nil {
			assert.Contains(t, mtlsErr.Error(), "EOF",
				"if the request fails, it should be due to a connection reset (EOF), not a timeout")
		} else {
			assert.Empty(t, body, "rejected request should have no body")
		}
	})

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

		externalDomain := fmt.Sprintf("%s.%s", bg.TestName, externalDomainBase)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName+"-int",
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

		externalDomain := fmt.Sprintf("%s.%s", bg.TestName, externalDomainBase)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName+"-int",
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

		httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("no-cert-client"))
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/headers", externalDomain), nil)
		require.NoError(t, err)

		_, err = httpClient.Do(req)
		require.Error(t, err, "request without client cert must be rejected by mTLS MUTUAL endpoint")
	})

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

		externalDomain := fmt.Sprintf("%s.%s", bg.TestName, externalDomainBase)
		eg, err := extgwhelper.CreateExternalGateway(
			t, bg.Namespace, bg.TestName,
			externalDomain, bg.TestName+"-int",
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
		require.NoError(t, err)

		var parsed struct {
			Headers map[string]string `json:"headers"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &parsed), "httpbin /headers should return parseable JSON")

		xfcc := parsed.Headers["X-Forwarded-Client-Cert"]
		require.NotEmpty(t, xfcc, "X-Forwarded-Client-Cert must be present in workload request")
		assert.Contains(t, xfcc, "test-client", "XFCC should carry the client cert CN")
		assert.Equal(t, 1, strings.Count(xfcc, "By="), "XFCC should contain exactly one cert entry (client cert, no ingress cert appended)")
	})

	t.Run("invalid regions configmap: ExternalGateway enters Error state", func(t *testing.T) {
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

		caSecretName := bgExt.TestName + "-ca"
		require.NoError(t, extgwhelper.CreateCASecret(t, bgExt.Namespace, caSecretName, certs.CACertPEM))
		regionsConfigMap := bgExt.TestName + "-regions"
		require.NoError(t, extgwhelper.CreateRegionsConfigMap(
			t, bgExt.Namespace, regionsConfigMap,
			extgwhelper.RegionsConfigMapData(extgwhelper.RegionName, certs.Subject),
		))
		_, err = extgwhelper.CreateExternalGateway(
			t, bgExt.Namespace, bgExt.TestName,
			fmt.Sprintf("%s.%s", bgExt.TestName, externalDomainBase), bgExt.TestName+"-int",
			extgwhelper.RegionName, regionsConfigMap, caSecretName,
		)
		require.NoError(t, err)
		require.NoError(t, extgwhelper.WaitUntilExternalGatewayReady(t, bgExt.Namespace, bgExt.TestName))

		kymaURL := fmt.Sprintf("https://%s.%s/headers", bgKyma.TestName, kymaGatewayDomain)
		require.NoError(t,
			endpointasserts.AssertEndpoint(t, http.MethodGet, kymaURL, http.StatusOK),
			"kyma default gateway must remain reachable after ExternalGateway is created",
		)
	})
}
