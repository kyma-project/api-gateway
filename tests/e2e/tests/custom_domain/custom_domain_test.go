package custom_domain

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	extgwhelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/extgateway"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
)

//go:embed custom_domain_no_auth_apirule.yaml
var APIRuleCustomDomainNoAuth string

//go:embed custom_domain_jwt_apirule.yaml
var APIRuleCustomDomainJWT string

//go:embed custom_domain_gateway.yaml
var customDomainGatewayTemplate string

// setupCustomGateway creates a custom Istio Gateway with a self-signed TLS certificate
// for the given host. It returns a "namespace/name" gateway reference for use in APIRule.
func setupCustomGateway(t *testing.T, namespace, name, host string) (string, error) {
	t.Helper()

	certPEM, keyPEM, err := extgwhelper.GenerateServerTLSCert(t, host)
	if err != nil {
		return "", fmt.Errorf("generating server TLS cert for %s: %w", host, err)
	}

	tlsSecretName := name + "-tls"
	if err := extgwhelper.CreateServerTLSSecret(t, tlsSecretName, certPEM, keyPEM); err != nil {
		return "", fmt.Errorf("creating TLS secret %s: %w", tlsSecretName, err)
	}

	_, err = infrahelpers.CreateResourceWithTemplateValues(
		t,
		customDomainGatewayTemplate,
		map[string]any{
			"Name":          name,
			"Host":          host,
			"TLSSecretName": tlsSecretName,
		},
		decoder.MutateNamespace(namespace),
	)
	if err != nil {
		return "", fmt.Errorf("creating custom Istio Gateway %s/%s: %w", namespace, name, err)
	}

	return fmt.Sprintf("%s/%s", namespace, name), nil
}

func TestAPIRuleCustomDomain(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	// Implements following tests/integration scenarios:
	//   - Scenario: Calling an unsecured API endpoint with custom domain
	t.Run("Calling an unsecured API endpoint with custom domain", func(t *testing.T) {
		t.Parallel()

		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("custom-domain-noauth"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		host := fmt.Sprintf("httpbin-%s.%s", testBackground.TestName, kymaGatewayDomain)

		gatewayRef, err := setupCustomGateway(t, testBackground.Namespace, testBackground.TestName, host)
		require.NoError(t, err, "Failed to create custom Istio Gateway")

		_, err = infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleCustomDomainNoAuth,
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

		statusCode, _, _, err := testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithoutToken())
		require.NoError(t, err, "Request without token should succeed for noAuth endpoint")
		assert.GreaterOrEqual(t, statusCode, http.StatusOK)
		assert.Less(t, statusCode, http.StatusMultipleChoices)

		statusCode, _, _, err = testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithTokenOverride("any-token"))
		require.NoError(t, err, "Request with arbitrary token should succeed for noAuth endpoint")
		assert.GreaterOrEqual(t, statusCode, http.StatusOK)
		assert.Less(t, statusCode, http.StatusMultipleChoices)
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Calling a secured API with JWT and custom domain
	t.Run("Calling a secured API with JWT and custom domain", func(t *testing.T) {
		t.Parallel()

		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("custom-domain-jwt"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		host := fmt.Sprintf("httpbin-%s.%s", testBackground.TestName, kymaGatewayDomain)

		gatewayRef, err := setupCustomGateway(t, testBackground.Namespace, testBackground.TestName, host)
		require.NoError(t, err, "Failed to create custom Istio Gateway")

		_, err = infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleCustomDomainJWT,
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

		statusCode, _, _, err := testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithoutToken())
		require.NoError(t, err, "Request without token should be rejected for JWT-secured endpoint")
		assert.GreaterOrEqual(t, statusCode, http.StatusBadRequest)
		assert.LessOrEqual(t, statusCode, http.StatusForbidden)

		statusCode, _, _, err = testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithTokenOverride("any-token"))
		require.NoError(t, err, "Request with invalid token should be rejected for JWT-secured endpoint")
		assert.GreaterOrEqual(t, statusCode, http.StatusBadRequest)
		assert.LessOrEqual(t, statusCode, http.StatusForbidden)

		validToken, err := testBackground.Provider.GetToken(t)
		require.NoError(t, err, "Fetching valid token should succeed")

		statusCode, _, _, err = testBackground.Provider.MakeRequest(t, http.MethodGet, url, oauth2.WithTokenOverride(validToken))
		require.NoError(t, err, "Request with valid JWT token should succeed")
		assert.GreaterOrEqual(t, statusCode, http.StatusOK)
		assert.Less(t, statusCode, http.StatusMultipleChoices)
	})
}
