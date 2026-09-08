package local_rate_limit_ingress_test

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
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

//go:embed ratelimit-api-rule.yaml
var RateLimitAPIRule string

//go:embed ratelimit-ingressgateway-default-bucket.yaml
var RateLimitIngressGatewayDefaultBucket string

//go:embed ratelimit-ingressgateway-path-based.yaml
var RateLimitIngressGatewayPathBased string

//go:embed ratelimit-ingressgateway-header-based.yaml
var RateLimitIngressGatewayHeaderBased string

//go:embed ratelimit-ingressgateway-path-and-header-based.yaml
var RateLimitIngressGatewayPathAndHeaderBased string

func TestLocalRateLimitIngress(t *testing.T) {
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t), "Failed to create APIGateway CR")

	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-igw"))
	require.NoError(t, err, "Failed to setup test namespace with httpbin")

	ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
		"TestID":           testBackground.TestName,
		"Namespace":        testBackground.Namespace,
		"GatewayNamespace": "kyma-system",
		"GatewayName":      "kyma-gateway",
		"Domain":           kymaGatewayDomain,
	}, testBackground.Namespace)

	baseURL := fmt.Sprintf("https://%s.%s", testBackground.TestName, kymaGatewayDomain)

	t.Run("Ingress gateway rate limited by default bucket", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-default", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayDefaultBucket, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, baseURL+"/ip", nil)
	})

	t.Run("Ingress gateway rate limited by path-based configuration", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-path", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, baseURL+"/ip", nil)
	})

	t.Run("Ingress gateway not rate limited by path-based configuration with wrong path", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-path-wp", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertNotRateLimited(t, http.MethodGet, baseURL+"/headers", nil, 5)
	})

	t.Run("Ingress gateway rate limited by header-based configuration", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-hdr", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayHeaderBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, baseURL+"/ip", map[string]string{"X-Rate-Limited": "true"})
	})

	t.Run("Ingress gateway not rate limited by header-based configuration with wrong header", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-hdr-wh", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayHeaderBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertNotRateLimited(t, http.MethodGet, baseURL+"/ip", map[string]string{"Different-Header": "true"}, 5)
	})

	t.Run("Ingress gateway rate limited by path and header based configuration", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-ph", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathAndHeaderBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, baseURL+"/headers", map[string]string{"X-Rate-Limited": "true"})
	})

	t.Run("Ingress gateway not rate limited by path and header based configuration with wrong path", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-ph-wp", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathAndHeaderBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertNotRateLimited(t, http.MethodGet, baseURL+"/ip", map[string]string{"X-Rate-Limited": "true"}, 5)
	})

	t.Run("Ingress gateway not rate limited by path and header based configuration with wrong header", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-ph-wh", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathAndHeaderBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertNotRateLimited(t, http.MethodGet, baseURL+"/headers", map[string]string{"Different-Header": "true"}, 5)
	})

	t.Run("Ingress gateway not rate limited by path and header based configuration with wrong path and wrong header", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-ph-wph", 16)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathAndHeaderBased, map[string]any{
			"Name": rlName,
		}, "istio-system")

		ratelimitasserts.AssertNotRateLimited(t, http.MethodGet, baseURL+"/ip", map[string]string{"Different-Header": "true"}, 5)
	})
}
