package asterisk

import (
	"fmt"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/endpoint"
	"sigs.k8s.io/e2e-framework/klient/decoder"

	_ "embed"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
)

//go:embed asterisk_paths.yaml
var APIRuleAsteriskPath string

func TestAPIRuleAsterisk(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("APIRule exposing service using asterisk in paths", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("asterisk"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleAsteriskPath,
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
			endpoint           string
			method             string
			expectedStatusCode int
		}{
			{endpoint: "/anything/one", method: http.MethodGet, expectedStatusCode: 200},
			{endpoint: "/anything/one/two", method: http.MethodGet, expectedStatusCode: 200},
			{endpoint: "/anything/random/one", method: http.MethodGet, expectedStatusCode: 200},
			{endpoint: "/anything/rand*m/one", method: http.MethodGet, expectedStatusCode: 200},
			{endpoint: "/anything/random/one/any/random/two", method: http.MethodDelete, expectedStatusCode: 200},
			{endpoint: "/anything/rand*m/one/any/rand*m/two", method: http.MethodDelete, expectedStatusCode: 200},
			{endpoint: "/anything/any/random/two", method: http.MethodGet, expectedStatusCode: 200},
			{endpoint: "/anything/any/random/two", method: http.MethodPost, expectedStatusCode: 200},
			{endpoint: "/anything/random/one", method: http.MethodPost, expectedStatusCode: 404},
			{endpoint: "/anything/", method: http.MethodGet, expectedStatusCode: 200},
			{endpoint: "/anything/a+b", method: http.MethodPut, expectedStatusCode: 200},
			{endpoint: "/anything/a%20b", method: http.MethodPut, expectedStatusCode: 200},
			{endpoint: "/anything/rand*m", method: http.MethodPut, expectedStatusCode: 200},
			{endpoint: "/anything/any/random", method: http.MethodPut, expectedStatusCode: 200},
		}

		for _, request := range requests {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, request.endpoint)
			err := endpoint.AssertEndpoint(t, request.method, url, request.expectedStatusCode)
			if err != nil {
				t.Fatalf("err %s", err.Error())
			}
		}
	})
}
