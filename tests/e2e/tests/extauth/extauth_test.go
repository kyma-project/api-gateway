package extauth

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	extauth "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/extauth"
	istioasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/istio"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	extauthhelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/extauth"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
)

//go:embed ext_auth_apirule.yaml
var APIRuleExtAuth string

//go:embed ext_auth_apirule_with_jwt.yaml
var APIRuleExtAuthJWT string

func TestAPIRuleExtAuth(t *testing.T) {

	require.NoError(t, modulehelpers.CreateDeprecatedV1configMap(t))
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	require.NoError(t, extauthhelper.CreateExtAuth(t))

	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	// Implements:
	// - Calling a httpbin endpoint secured with ExtAuth
	t.Run("access to extAuth exposed service with no authorization", func(t *testing.T) {
		t.Parallel()
		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("extauth-test"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdAPIRule, err := infrastructure.CreateResourceWithTemplateValues(
			t,
			APIRuleExtAuth,
			map[string]any{
				"Name":            testBackground.TestName,
				"Host":            testBackground.TestName,
				"ServiceName":     testBackground.TargetServiceName,
				"ServicePort":     testBackground.TargetServicePort,
				"Gateway":         "kyma-system/kyma-gateway",
				"ExtAuthProvider": "sample-ext-authz-http",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdAPIRule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		// then
		err = extauth.AssertEndpoint(t, http.MethodGet, fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain), map[string]string{"x-ext-authz": "deny"}, 403)
		require.NoError(t, err, "Request should be forbidden without valid token")

		err = extauth.AssertEndpoint(t, http.MethodGet, fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain), map[string]string{"x-ext-authz": "allow"}, 200)
		require.NoError(t, err, "Request should be allowed with valid token")

	})

	// - Calling a httpbin endpoint secured with ExtAuth with JWT restrictions
	t.Run("access to extAuth exposed service with JWT restrictions", func(t *testing.T) {
		t.Parallel()
		// given
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("extauth-jwt"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdAPIRule, err := infrastructure.CreateResourceWithTemplateValues(
			t,
			APIRuleExtAuthJWT,
			map[string]any{
				"Name":            testBackground.TestName,
				"Host":            testBackground.TestName,
				"ServiceName":     testBackground.TargetServiceName,
				"ServicePort":     testBackground.TargetServicePort,
				"IssuerUrl":       testBackground.Provider.GetIssuerURL(),
				"JwksUri":         testBackground.Provider.GetJwksURI(),
				"Gateway":         "kyma-system/kyma-gateway",
				"ExtAuthProvider": "sample-ext-authz-http",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdAPIRule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		istioasserts.VirtualServiceOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace)
		istioasserts.AuthorizationPolicyOwnedByAPIRuleExists(t, testBackground.Namespace, testBackground.TestName, testBackground.Namespace, 2)

		// then
		//Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and no token should result in status 403
		err = extauth.AssertEndpointWithJWT(t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain),
			403,
			testBackground.Provider,
			oauth2.WithHeaders(map[string]string{
				"x-ext-authz": "allow",
			}),
			oauth2.WithoutToken(),
		)
		require.NoError(t, err, "Request should be forbidden without valid JWT token")

		//And Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and an invalid "JWT" token should result in status between 400
		err = extauth.AssertEndpointWithJWT(t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain),
			401,
			testBackground.Provider,
			oauth2.WithHeaders(
				map[string]string{
					"x-ext-authz": "allow",
				}),
			oauth2.WithTokenOverride("invalid-token"),
		)
		require.NoError(t, err, "Request should be forbidden with invalid JWT token")

		//And Calling the "/headers" endpoint with header "x-ext-authz" with value "deny" and a valid "JWT" token should result in status between 400 and 403
		err = extauth.AssertEndpointWithJWT(t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain),
			403,
			testBackground.Provider,
			oauth2.WithHeaders(
				map[string]string{
					"x-ext-authz": "deny",
				}),
		)
		require.NoError(t, err, "Request should be forbidden with invalid token")

		//And Calling the "/headers" endpoint with header "x-ext-authz" with value "allow" and a valid "JWT" token should result in status between 200 and 299
		err = extauth.AssertEndpointWithJWT(t,
			http.MethodGet,
			fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain),
			200,
			testBackground.Provider,
			oauth2.WithHeaders(
				map[string]string{
					"x-ext-authz": "allow",
				}),
		)
		require.NoError(t, err, "Request should be allowed with valid token")
	})
}
