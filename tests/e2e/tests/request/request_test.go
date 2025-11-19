package request

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	h "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed jwt_request_header.yaml
var APIRuleRequestHeader string

//go:embed jwt_request_cookie.yaml
var APIRuleRequestCookie string

func TestAPIRuleRequestHeadersAndCookies(t *testing.T) {
	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("Exposing an endpoint with request header configured", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("requests"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleRequestHeader,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"HeaderKey":   "X-Request-Test",
				"HeaderValue": "a-header-value",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/headers")
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}
		c := h.NewHTTPClient(t)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}
		require.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}

		responseBody := map[string]map[string][]string{}

		err = json.Unmarshal(body, &responseBody)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}

		headers := responseBody["headers"]
		assert.Equal(t, "a-header-value", headers["X-Request-Test"][0])
	})

	t.Run("Exposing an endpoint with request cookie configured", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("requests"))
		require.NoError(t, err, "Failed to setup test background with httpbin")

		createdApirule, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			APIRuleRequestCookie,
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"Cookie":      "x-request-test",
				"CookieValue": "a-cookie-value",
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s%s", testBackground.TestName, kymaGatewayDomain, "/headers")
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}
		c := h.NewHTTPClient(t)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}
		require.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}

		responseBody := map[string]map[string][]string{}

		err = json.Unmarshal(body, &responseBody)
		if err != nil {
			t.Fatalf("err %s", err.Error())
		}

		headers := responseBody["headers"]
		assert.Equal(t, "x-request-test=a-cookie-value", headers["Cookie"][0])
	})
}
