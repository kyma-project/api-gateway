package cors

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
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed cors_default.yaml
var APIRuleCorsDefault string

//go:embed cors_custom.yaml
var APIRuleCorsCustom string

func TestAPIRuleCors(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("No CORS headers are returned when CORS is not specified in the APIRule", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("cors"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleCorsDefault,
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
		absentHeaders := []string{
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Methods",
			"Access-Control-Allow-Headers",
			"Access-Control-Expose-Headers",
			"Access-Control-Allow-Credentials",
			"Access-Control-Max-Age",
		}
		err = endpoint.AssertEndpointWithoutResponseHeaders(t, http.MethodOptions, url, map[string]string{"Origin": "localhost"}, http.StatusOK, absentHeaders)
		require.NoError(t, err, "Failed to make http request with CORS")
	})

	t.Run("CORS headers are returned when CORS is specified in the APIRule", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("cors"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleCorsCustom,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"AllowOrigins": []map[string]string{
					{"regex": ".*local.kyma.dev"},
				},
				"AllowMethods":     []string{"GET", "POST"},
				"AllowHeaders":     []string{"x-custom-allow-headers"},
				"AllowCredentials": "false",
				"ExposeHeaders":    []string{"x-custom-expose-headers"},
				"MaxAge":           "300",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		allowedOriginHeaderValues := []string{
			"test.local.kyma.dev",
			"a.local.kyma.dev",
			"b.local.kyma.dev",
			"c.local.kyma.dev",
			"d.local.kyma.dev",
		}

		for _, originHeaderValue := range allowedOriginHeaderValues {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")
			requestHeaders := map[string]string{
				"Origin":                        originHeaderValue,
				"Access-Control-Request-Method": "GET,POST,PUT,DELETE,PATCH",
			}
			expectedResponseHeaders := map[string]string{
				"Access-Control-Allow-Origin":   originHeaderValue,
				"Access-Control-Allow-Methods":  "GET,POST",
				"Access-Control-Allow-Headers":  "x-custom-allow-headers",
				"Access-Control-Expose-Headers": "x-custom-expose-headers",
				"Access-Control-Max-Age":        "300",
			}
			err = endpoint.AssertEndpointWithResponseHeaders(t, http.MethodOptions, url, requestHeaders, http.StatusOK, expectedResponseHeaders)
			require.NoError(t, err, "Failed to make http request with CORS")
		}

		notAllowedOriginHeaderValues := []string{
			"localhost",
			"a.localhost",
			"a.b.localhost",
		}

		for _, originHeaderValue := range notAllowedOriginHeaderValues {
			url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/ip")
			requestHeaders := map[string]string{
				"Origin":                        originHeaderValue,
				"Access-Control-Request-Method": "GET,POST,PUT,DELETE,PATCH",
			}
			absentHeaders := []string{
				"Access-Control-Allow-Origin",
				"Access-Control-Allow-Methods",
				"Access-Control-Allow-Headers",
				"Access-Control-Expose-Headers",
				"Access-Control-Allow-Credentials",
				"Access-Control-Max-Age",
			}
			err = endpoint.AssertEndpointWithoutResponseHeaders(t, http.MethodOptions, url, requestHeaders, http.StatusOK, absentHeaders)
			require.NoError(t, err, "Failed to make http request with CORS")
		}
	})
}
