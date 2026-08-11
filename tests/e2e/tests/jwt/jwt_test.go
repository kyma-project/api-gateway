package jwt

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	istioasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/istio"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpincluster"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
)

//go:embed jwt_apirule.yaml
var APIRuleJWT string

//go:embed jwt_apirule_with_scope.yaml
var APIRuleJWTWithScope string

//go:embed jwt_and_unrestricted.yaml
var APIRuleJWTMixed string

//go:embed jwt_apirule_with_audience.yaml
var APIRuleJWTWithAudience string

//go:embed jwt_apirule_token_from_header.yaml
var APIRuleJWTFromHeader string

//go:embed jwt_apirule_token_from_param.yaml
var APIRuleJWTFromParam string

//go:embed jwt_apirule_unavailable_issuer.yaml
var APIRuleJWTUnavailableIssuer string

//go:embed jwt_apirule_issuer_jwks_not_match.yaml
var APIRuleJWTIssuerNotMatchingJwks string

func TestAPIRuleJWT(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	// Implements following tests/integration scenarios:
	//   - Scenario: Calling a httpbin endpoint secured
	//   - Scenario: Calling a httpbin endpoint secured on all paths
	//   - Scenario: In-cluster calling a httpbin endpoint secured with JWT
	t.Run("access to JWT exposed service with no authorizations", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
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

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 1)

		// then
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/123", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
		)
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/goat", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
		)

		// in-cluster call
		stdOut, stdErr, err := httpincluster.RunRequestFromInsideCluster(t,
			testBackground.Namespace,
			fmt.Sprintf("http://%s.%s.svc.cluster.local:%d/anything/cluster",
				testBackground.TargetServiceName, testBackground.Namespace, testBackground.TargetServicePort,
			),
		)

		assert.Error(t, err, "Expected error when calling secured endpoint without token")
		assert.NotEmpty(t, stdOut, "StdOut should not be empty")
		assert.NotEmpty(t, stdErr, "StdErr should not be empty")
		assert.Contains(t, stdOut, "HTTP/1.1 403 Forbidden", "Response should contain 403 Forbidden")
		assert.Contains(t, stdErr, "The requested URL returned error: 403")
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Calling httpbin that has an endpoint secured by JWT and unrestricted endpoint
	t.Run("access to path exposed with JWT and different exposed with noAuth", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-mixed"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleJWTMixed,
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
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		// then
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/jwt", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
		)

		code, headers, _, err := testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/noauth", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithoutToken(),
		)
		require.NoError(t, err, "Failed to make request without token")
		assert.Equal(t, http.StatusOK, code, "Response should be 200")
		assert.Contains(t, headers, "X-Envoy-Upstream-Service-Time", "Response headers should contain X-Envoy-Upstream-Service-Time")
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Calling a httpbin endpoint secured with JWT that requires scopes claims
	t.Run("access to JWT exposed service that requires JWT scope claims read and write", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-scope"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
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
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource with scope")
		require.NotEmpty(t, createdApirule, "Created APIRule resource with scope should not be empty")

		// then
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/readwrite", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
			oauth2.WithGetTokenOptions(
				oauth2.WithScope("read"),
				oauth2.WithScope("write"),
			),
		)

		code, _, _, err := testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/readwrite", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithGetTokenOption(oauth2.WithScope("read")),
		)
		require.NoError(t, err, "Failed to make request with insufficient scope")
		assert.Equal(t, http.StatusForbidden, code, "Response should be 403 for insufficient scope")

		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/read", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
			oauth2.WithGetTokenOption(oauth2.WithScope("read")),
		)

		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/read", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
			oauth2.WithGetTokenOptions(
				oauth2.WithScope("read"),
				oauth2.WithScope("write"),
			),
		)
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Calling a httpbin endpoint secured with JWT that requires aud claim
	t.Run("access to JWT exposed service that requires JWT aud claim", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-aud"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleJWTWithAudience,
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
		require.NoError(t, err, "Failed to create APIRule resource with audience")
		require.NotEmpty(t, createdApirule, "Created APIRule resource with audience should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		// then
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/goats", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
			oauth2.WithGetTokenOption(oauth2.WithAudience("goats")),
		)

		code, _, _, err := testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/goats", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithGetTokenOption(oauth2.WithAudience("sheeps")),
		)
		require.NoError(t, err, "Failed to make request with wrong audience")
		assert.Equal(t, http.StatusForbidden, code, "Response should be 403 for wrong audience")

		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/goatsheeps", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
			oauth2.WithGetTokenOptions(
				oauth2.WithAudience("goats"),
				oauth2.WithAudience("sheeps"),
			),
		)

		code, _, _, err = testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/goatsheeps", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithGetTokenOption(oauth2.WithAudience("goats")),
		)
		require.NoError(t, err, "Failed to make request with insufficient audience")
		assert.Equal(t, http.StatusForbidden, code, "Response should be 403 for insufficient audience")
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Exposing an endpoint secured with different JWT token from headers
	t.Run("access to JWT exposed service with token from custom header", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-header"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleJWTFromHeader,
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
		require.NoError(t, err, "Failed to create APIRule resource with token from header")
		require.NotEmpty(t, createdApirule, "Created APIRule resource with token from header should not be empty")
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 1)

		// then
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/header", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
			oauth2.WithTokenHeader("Custom-Token-Header"),
			oauth2.WithTokenPrefix("JwtToken"),
		)

		code, _, _, err := testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/header", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithTokenHeader("Another-Custom-Token-Header"),
			oauth2.WithTokenPrefix("JwtToken"),
		)

		assert.NoError(t, err, "Failed to make request with token from different header")
		assert.Equal(t, http.StatusForbidden, code, "Response should be 403 for token from different header")

		code, _, _, err = testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/header", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithTokenHeader("Authorization"),
			oauth2.WithTokenPrefix("Bearer"),
		)
		assert.NoError(t, err, "Failed to make request with token from Authorization header")
		assert.Equal(t, http.StatusForbidden, code, "Response should be 403 for token from Authorization header")
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Calling a httpbin endpoint secured with different JWT token from params
	t.Run("access to JWT exposed service with token from params", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-params"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleJWTFromParam,
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
		require.NoError(t, err, "Failed to create APIRule resource with token from params")
		require.NotEmpty(t, createdApirule, "Created APIRule resource with token from params should not be empty")
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 1)

		// then
		oauth2.AssertEndpointWithProvider(
			t,
			testBackground.Provider,
			fmt.Sprintf("https://%s.%s/anything/param", testBackground.TestName, kymaGatewayDomain),
			http.MethodGet,
			oauth2.WithTokenFromParam("token_param"),
		)

		code, _, _, err := testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/param", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithTokenFromParam("different_param"),
		)
		assert.NoError(t, err, "Failed to make request with token from different param")
		assert.Equal(t, http.StatusForbidden, code, "Response should be 403 for token from different param")

		code, _, _, err = testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/param", testBackground.TestName, kymaGatewayDomain),
			oauth2.WithTokenHeader("Authorization"),
		)
		assert.NoError(t, err, "Failed to make request with token from Authorization param")
		assert.Equal(t, http.StatusForbidden, code, "Response should be 403 for token from Authorization param")
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Exposing a JWT secured endpoint with unavailable issuer and jwks URL
	t.Run("access to JWT exposed service with not existing issuer and jwks", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-bad-issuer"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleJWTUnavailableIssuer,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)

		require.NoError(t, err, "Failed to create APIRule resource with token from params")
		require.NotEmpty(t, createdApirule, "Created APIRule resource with token from params should not be empty")
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 1)

		// then
		code, _, body, err := testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/unavailable-issuer", testBackground.TestName, kymaGatewayDomain),
		)
		assert.NoError(t, err, "Failed to make request with token to endpoint with unavailable issuer")
		assert.Equal(t, http.StatusUnauthorized, code, "Response should be 401 for unavailable issuer")
		assert.Contains(t, string(body), "Jwt issuer is not configured", "Response body should contain invalid signature error")
	})

	// Implements following tests/integration scenarios:
	//   - Scenario: Exposing a JWT secured endpoint where issuer URL doesn't belong to jwks URL
	t.Run("access to JWT exposed service with not matching issuer and jwks", func(t *testing.T) {
		t.Parallel()

		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("jwt-test-bad-issuer"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleJWTIssuerNotMatchingJwks,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"Issuer":      testBackground.Provider.GetIssuerURL(),
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)

		require.NoError(t, err, "Failed to create APIRule resource with token from params")
		require.NotEmpty(t, createdApirule, "Created APIRule resource with token from params should not be empty")
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 1)

		// then
		code, _, body, err := testBackground.Provider.MakeRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/anything/issuer-jwks-not-match", testBackground.TestName, kymaGatewayDomain),
		)
		assert.NoError(t, err, "Failed to make request with token to endpoint with unavailable issuer")
		assert.Equal(t, http.StatusUnauthorized, code, "Response should be 401 for unavailable issuer")
		assert.Contains(t, string(body), "Jwt verification fails", "Response body should contain JWT verification failure error")
	})
}
