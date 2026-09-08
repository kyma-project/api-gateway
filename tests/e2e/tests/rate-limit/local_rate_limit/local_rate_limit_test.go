package local_rate_limit

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	ratelimitasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/ratelimit"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
)

//go:embed ratelimit-api-rule.yaml
var RateLimitAPIRule string

//go:embed ratelimit-default-bucket.yaml
var RateLimitDefaultBucket string

//go:embed ratelimit-path-based.yaml
var RateLimitPathBased string

//go:embed ratelimit-header-based.yaml
var RateLimitHeaderBased string

//go:embed ratelimit-path-and-header-based.yaml
var RateLimitPathAndHeaderBased string

func TestLocalRateLimit(t *testing.T) {
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t), "Failed to create APIGateway CR")

	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("Pod rate limited by default bucket", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-default"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitDefaultBucket, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, nil)
	})

	t.Run("Pod rate limited by path-based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-path"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitPathBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, nil)
	})

	t.Run("Pod rate limited by header-based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-header"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitHeaderBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, map[string]string{"X-Rate-Limited": "true"})
	})

	t.Run("Pod rate limited by path and header based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-path-hdr"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitPathAndHeaderBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, map[string]string{"X-Rate-Limited": "true"})
	})
}
