package no_auth

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/endpoint"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpincluster"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed no_auth_wildcard.yaml
var APIRuleNoAuthWildcard string

//go:embed no_auth_wildcard_updated.yaml
var APIRuleNoAuthWildcardUpdated string

func TestAPIRuleNoAuth(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	t.Run("Calling an endpoint unsecured on all paths from outside of the cluster", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("no-auth"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcard,
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
			{path: "/ip", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
		}

		for _, r := range requests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			err = endpoint.AssertEndpoint(t, r.method, url, r.expectedResponseStatus)
			require.NoError(t, err)
		}
	})

	t.Run("Calling an endpoint unsecured on all paths from inside of the cluster", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("no-auth"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcard,
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

		requestPaths := []string{
			"/status/200",
			"/headers",
		}

		for _, path := range requestPaths {
			// in-cluster call
			stdOut, stdErr, err := httpincluster.RunRequestFromInsideCluster(t,
				testBackground.Namespace,
				fmt.Sprintf("http://%s.%s.svc.cluster.local:%d%s",
					testBackground.TargetServiceName, testBackground.Namespace, testBackground.TargetServicePort, path,
				),
			)

			assert.Error(t, err, "Expected error when calling another service within the mesh")
			assert.NotEmpty(t, stdOut, "StdOut should not be empty")
			assert.NotEmpty(t, stdErr, "StdErr should not be empty")
			assert.Contains(t, stdOut, "HTTP/1.1 403 Forbidden", "Response should contain 403 Forbidden")
			assert.Contains(t, stdErr, "The requested URL returned error: 403")

		}
	})

	t.Run("Updating an APIRule and calling an httpbin endpoint unsecured on all paths", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("no-auth"))
		require.NoError(t, err, "Failed to setup test background with httpbin")
		kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
		require.NoError(t, err, "Failed to get domain from kyma-gateway")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcard,
			// got to fulfill these properly
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

		updatedApirule, err := infrahelpers.UpdateResourceWithTemplateValues(
			t,
			APIRuleNoAuthWildcardUpdated,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)

		require.NoError(t, err, "Failed to update APIRule resource")
		require.NotEmpty(t, updatedApirule, "Updated APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		requests := []struct {
			path                   string
			method                 string
			expectedResponseStatus int
		}{
			{path: "/ip", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/status/200", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
			{path: "/headers", method: http.MethodGet, expectedResponseStatus: http.StatusOK},
		}

		for _, r := range requests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, r.path)
			err = endpoint.AssertEndpoint(t, r.method, url, r.expectedResponseStatus)
			require.NoError(t, err)
		}
	})
}
