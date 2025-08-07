package jwt

import (
	_ "embed"
	"fmt"
	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	istioasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/istio"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/assert"
	"net/http"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed jwt_apirule.yaml
var APIRuleJWT string

//go:embed jwt_apirule_with_scope.yaml
var APIRuleJWTWithScope string

func TestAPIRuleJWT(t *testing.T) {
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("access to JWT exposed service with no authorizations", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrastructure.CreateResourceWithTemplateValues(
			t,
			APIRuleJWT,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"Issuer":      testBackground.Provider.GetIssuerURL(),
				"JwksUri":     testBackground.Provider.GetJwksURI(),
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		// then
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)

		code, headers, body, err := testBackground.Provider.MakeRequestWithToken(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain),
		)
		require.NoError(t, err, "Failed to make request with mock token")
		assert.Equal(t, http.StatusOK, code, "Response should be 200")
		assert.Contains(t, headers, "X-Envoy-Upstream-Service-Time", "Response headers should contain X-Envoy-Upstream-Service-Time")
		assert.Contains(t, string(body), "Authorization", "Response body should contain Authorization header")
	})

	t.Run("access to JWT exposed service with scope", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-scope"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrastructure.CreateResourceWithTemplateValues(
			t,
			APIRuleJWTWithScope,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"Issuer":      testBackground.Provider.GetIssuerURL(),
				"JwksUri":     testBackground.Provider.GetJwksURI(),
				"Scope":       "read",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource with scope")
		require.NotEmpty(t, createdApirule, "Created APIRule resource with scope should not be empty")

		// then
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)

		code, headers, body, err := testBackground.Provider.MakeRequestWithToken(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithScope("read"),
		)

		require.NoError(t, err, "Failed to make request with mock token")
		assert.Equal(t, http.StatusOK, code, "Response should be 200")
		assert.Contains(t, headers, "X-Envoy-Upstream-Service-Time", "Response headers should contain X-Envoy-Upstream-Service-Time")
		assert.Contains(t, string(body), "Authorization", "Response body should contain Authorization header")
	})
}
