package jwt

import (
	_ "embed"
	v2 "github.com/kyma-project/api-gateway/apis/gateway/v2"
	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testid"
	"github.com/stretchr/testify/assert"

	"net/http"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed jwt_apirule.yaml
var APIRuleJWT string

//go:embed jwt_apirule_with_scope.yaml
var APIRuleJWTWithScope []byte

func TestAPIRuleJWT(t *testing.T) {
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	// TODO: karteczka na tracing tworzenia zasobów z APIServera
	// TODO: kopiowanie logów z podów w trakcie (przed) cleanupem
	t.Run("access to JWT exposed service with no authorizations", func(t *testing.T) {
		// given

		// TODO: change testid package name
		testBackground, err := testid.SetupRandomNamespaceWithOauth2MockAndHttpbin(t, testid.WithPrefix("jwt-test"))
		require.NoError(t, err, "Failed to setup test background with OAuth2 mock and httpbin")

		// when
		createdApirule, err := infrastructure.CreateResourceWithTemplateValues(
			t,
			APIRuleJWT,
			&v2.APIRule{},
			map[string]any{
				"Name":        testBackground.TestName,
				"Host":        testBackground.TestName,
				"ServiceName": testBackground.TargetServiceName,
				"ServicePort": testBackground.TargetServicePort,
				"Gateway":     "kyma-system/kyma-gateway",
				"Issuer":      testBackground.Mock.IssuerURL,
				"JwksUri":     testBackground.Mock.JwksURI,
			},
			decoder.MutateNamespace(testBackground.Namespace),
		)
		require.NoError(t, err, "Failed to create APIRule resource")
		require.NotEmpty(t, createdApirule, "Created APIRule resource should not be empty")

		// then
		apiruleasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)
		// TODO:
		//istioAsserts.VSExist(t, testBackground.TestName, testBackground.Namespace)
		//istioAsserts.APExist(t, testBackground.TestName, testBackground.Namespace)
		// maybe: check confguration of Istio sidecar via admin API (proxy config dump)

		// TODO: discover domain from kyma-gateway
		// TODO: add readiness check for oauth2-mock and httpbin
		code, headers, body, err := testBackground.Mock.MakeRequestWithMockToken(t, http.MethodGet, "https://"+testBackground.TestName+".local.kyma.dev/headers")
		require.NoError(t, err, "Failed to make request with mock token")
		assert.Equal(t, http.StatusOK, code, "Response should be 200")
		assert.Contains(t, headers, "X-Envoy-Upstream-Service-Time", "Response headers should contain X-Envoy-Upstream-Service-Time")
		assert.Contains(t, string(body), "Authorization", "Response body should contain Authorization header")
	})
}
