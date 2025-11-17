package short_host

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/endpoint"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed short_host.yaml
var APIRuleShortHost string

func TestAPIRuleShortHost(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("Exposing an unsecured service with short host", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("short-host"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleShortHost,
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

		url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")
		err = endpoint.AssertEndpoint(t, http.MethodGet, url, http.StatusOK)
		require.NoError(t, err)
	})

	t.Run("Exposing an unsecured service with short host (error scenario)", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("short-host"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleShortHost,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/another-gateway",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilError(t, testBackground.TestName, testBackground.Namespace)

		isErr := apiruleasserts.HasState(t, testBackground.TestName, testBackground.Namespace, v2.Error)
		require.True(t, isErr)

		isDescCorrect := apiruleasserts.HasStatusDescription(t, testBackground.TestName, testBackground.Namespace, "Validation errors: Attribute 'spec.gateway': Could not get specified Gateway")
		require.True(t, isDescCorrect)
	})

}
