package expose_methods_on_paths

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/endpoint"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed paths_and_methods_noauth.yaml
var APIRuleNoAuth string

//go:embed paths_and_methods_jwt.yaml
var APIRuleJwt string

func TestAPIRuleMethodsOnPaths(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("Expose GET, POST method for /anything and only PUT for /anything/put with noAuth", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("methods-on-paths"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuth,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		requests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/anything", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/anything", method: http.MethodPost, expectedResponseStatus: http.StatusOK},
			{path: "/anything", method: http.MethodPut, expectedResponseStatus: http.StatusNotFound},
			{path: "/anything/put", method: http.MethodPut, expectedResponseStatus: http.StatusOK},
			{path: "/anything/put", method: http.MethodPost, expectedResponseStatus: http.StatusNotFound},
		}

		for _, r := range requests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			err = endpoint.AssertEndpoint(t, r.method, url, r.expectedResponseStatus)
			require.NoError(t, err, "Failed to make http request")
		}
	})

	t.Run("Expose GET, POST method for /anything and PUT for /anything/put secured by JWT", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testsetup.WithPrefix("methods-on-paths"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleJwt,
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

		exposedPathsRequests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/anything", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/anything", method: http.MethodPost, expectedResponseStatus: http.StatusOK},
			{path: "/anything/put", method: http.MethodPut, expectedResponseStatus: http.StatusOK},
		}

		for _, r := range exposedPathsRequests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			oauth2.AssertEndpointWithProvider(
				t,
				testBackground.Provider,
				url,
				r.method,
			)
		}

		nonExposedPathsRequests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/anything", method: http.MethodPut, expectedResponseStatus: http.StatusNotFound},
			{path: "/anything/put", method: http.MethodPost, expectedResponseStatus: http.StatusNotFound},
		}

		for _, r := range nonExposedPathsRequests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			oauth2.AssertNonExposedEndpointWithProvider(
				t,
				testBackground.Provider,
				url,
				r.method,
			)
		}
	})
}
