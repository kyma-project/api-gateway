package local_rate_limit_ingress

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
		rateLimitedPath := "/ip"
		maxTokens := 4
		tokensPerFill := 4

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayDefaultBucket, map[string]any{
			"Name":          rlName,
			"MaxTokens":     maxTokens,
			"TokensPerFill": tokensPerFill,
			"FillInterval":  "1h",
		}, "istio-system")

		pods := ratelimitasserts.MatchingPodCount(t, rlName, "istio-system")
		url := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		// With a single pod the token bucket is local and assertions are deterministic.
		// With multiple pods each pod has its own bucket; WithRetries distributes requests
		// across all pods to account for load balancer skew.
		if pods == 1 {
			// Assert 'maxTokens' successful responses based on the number of tokens in the default bucket.
			ratelimitasserts.AssertNSuccessfulResponses(t, maxTokens, http.MethodGet, url, nil)
			// Assert the next request is rate-limited because the default bucket is exhausted.
			ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)
			return
		}
		// Assert aggregate successful responses across per-pod default buckets.
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*maxTokens, pods*maxTokens*10, http.MethodGet, url, nil)
		// Assert additional requests are rate-limited once aggregate pod capacity is exhausted.
		ratelimitasserts.AssertNRateLimitedResponses(t, pods*10, http.MethodGet, url, nil)
	})

	t.Run("Ingress gateway rate limited by path-based configuration", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-path", 16)
		rateLimitedPath := "/ip"
		nonRateLimitedPath := "/headers"
		defaultMaxTokens := 8
		defaultTokensPerFill := 8
		pathMaxTokens := 4
		pathTokensPerFill := 4

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathBased, map[string]any{
			"Name":              rlName,
			"MaxTokens":         defaultMaxTokens,
			"TokensPerFill":     defaultTokensPerFill,
			"FillInterval":      "1h",
			"Path":              rateLimitedPath,
			"PathMaxTokens":     pathMaxTokens,
			"PathTokensPerFill": pathTokensPerFill,
			"PathFillInterval":  "1h",
		}, "istio-system")

		pods := ratelimitasserts.MatchingPodCount(t, rlName, "istio-system")
		urlRateLimited := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		urlDefaultBucket := fmt.Sprintf("%s%s", baseURL, nonRateLimitedPath)
		if pods == 1 {
			// Assert path-specific bucket limits traffic on the configured path.
			ratelimitasserts.AssertNSuccessfulResponses(t, pathMaxTokens, http.MethodGet, urlRateLimited, nil)
			ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlRateLimited, nil)
			// Assert default bucket limits traffic on paths without path-specific rules.
			ratelimitasserts.AssertNSuccessfulResponses(t, defaultMaxTokens, http.MethodGet, urlDefaultBucket, nil)
			ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlDefaultBucket, nil)
			return
		}
		// Assert aggregate path-specific capacity across all ingress gateway pods.
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*pathMaxTokens, pods*pathMaxTokens*10, http.MethodGet, urlRateLimited, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, pods*10, http.MethodGet, urlRateLimited, nil)
		// Assert aggregate default-bucket capacity on non-matching paths across all pods.
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*defaultMaxTokens, pods*defaultMaxTokens*10, http.MethodGet, urlDefaultBucket, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, pods*10, http.MethodGet, urlDefaultBucket, nil)
	})

	t.Run("Ingress gateway rate limited by header-based configuration", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-hdr", 16)
		rateLimitedPath := "/ip"
		rateLimitedHeaders := map[string]string{"X-Rate-Limited": "true"}
		nonRateLimitedHeaders := map[string]string{"Different-Header": "true"}
		defaultMaxTokens := 8
		defaultTokensPerFill := 8
		headerMaxTokens := 4
		headerTokensPerFill := 4

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayHeaderBased, map[string]any{
			"Name":                rlName,
			"MaxTokens":           defaultMaxTokens,
			"TokensPerFill":       defaultTokensPerFill,
			"FillInterval":        "1h",
			"Header":              "X-Rate-Limited",
			"HeaderValue":         "true",
			"HeaderMaxTokens":     headerMaxTokens,
			"HeaderTokensPerFill": headerTokensPerFill,
			"HeaderFillInterval":  "1h",
		}, "istio-system")

		pods := ratelimitasserts.MatchingPodCount(t, rlName, "istio-system")
		url := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		if pods == 1 {
			// Assert header-specific bucket applies when the configured header is present.
			ratelimitasserts.AssertNSuccessfulResponses(t, headerMaxTokens, http.MethodGet, url, rateLimitedHeaders)
			ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, rateLimitedHeaders)
			// Assert requests without the configured header are governed by the default bucket.
			ratelimitasserts.AssertNSuccessfulResponses(t, defaultMaxTokens, http.MethodGet, url, nonRateLimitedHeaders)
			ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nonRateLimitedHeaders)
			return
		}
		// Assert aggregate header-specific capacity across all ingress gateway pods.
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*headerMaxTokens, pods*headerMaxTokens*10, http.MethodGet, url, rateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, pods*10, http.MethodGet, url, rateLimitedHeaders)
		// Assert aggregate default-bucket capacity for non-matching headers across all pods.
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*defaultMaxTokens, pods*defaultMaxTokens*10, http.MethodGet, url, nonRateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, pods*10, http.MethodGet, url, nonRateLimitedHeaders)

	})

	t.Run("Ingress gateway rate limited by path and header based configuration", func(t *testing.T) {
		rlName := envconf.RandomName("rl-igw-ph", 16)
		rateLimitedPath := "/headers"
		nonRateLimitedPath := "/ip"
		rateLimitedHeaders := map[string]string{"X-Rate-Limited": "true"}
		nonRateLimitedHeaders := map[string]string{"Different-Header": "true"}
		defaultMaxTokens := 12
		defaultTokensPerFill := 12
		specializedMaxTokens := 4
		specializedTokensPerFill := 4

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathAndHeaderBased, map[string]any{
			"Name":                rlName,
			"MaxTokens":           defaultMaxTokens,
			"TokensPerFill":       defaultTokensPerFill,
			"FillInterval":        "1h",
			"Path":                rateLimitedPath,
			"Header":              "X-Rate-Limited",
			"HeaderValue":         "true",
			"HeaderMaxTokens":     specializedMaxTokens,
			"HeaderTokensPerFill": specializedTokensPerFill,
			"HeaderFillInterval":  "1h",
		}, "istio-system")

		pods := ratelimitasserts.MatchingPodCount(t, rlName, "istio-system")
		urlRateLimited := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		urlDefaultBucket := fmt.Sprintf("%s%s", baseURL, nonRateLimitedPath)
		if pods == 1 {
			// Assert combined path+header matching is rate limited by the specialized bucket.
			ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlRateLimited, rateLimitedHeaders)
			ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlRateLimited, rateLimitedHeaders)

			// Assert all non-matching combinations consume tokens from the default bucket.
			ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlDefaultBucket, rateLimitedHeaders)
			ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlRateLimited, nonRateLimitedHeaders)
			ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlDefaultBucket, nonRateLimitedHeaders)
			ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlDefaultBucket, nonRateLimitedHeaders)
			return
		}
		// Assert aggregate specialized bucket capacity for matching path+header traffic.
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*specializedMaxTokens, pods*specializedMaxTokens*10, http.MethodGet, urlRateLimited, rateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, pods*10, http.MethodGet, urlRateLimited, rateLimitedHeaders)
		// Assert non-matching combinations consume aggregate capacity from default buckets.
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*specializedMaxTokens, pods*specializedMaxTokens*10, http.MethodGet, urlDefaultBucket, rateLimitedHeaders)
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*specializedMaxTokens, pods*specializedMaxTokens*10, http.MethodGet, urlRateLimited, nonRateLimitedHeaders)
		ratelimitasserts.AssertNSuccessfulResponsesWithRetries(t, pods*specializedMaxTokens, pods*specializedMaxTokens*10, http.MethodGet, urlDefaultBucket, nonRateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, pods*10, http.MethodGet, urlDefaultBucket, nonRateLimitedHeaders)

	})

	t.Run("Ingress gateway response headers present when enableResponseHeaders is true", func(t *testing.T) {
		maxTokens := 4
		tokensPerFill := 4
		rlName := envconf.RandomName("rl-igw-resp-hdr", 24)
		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayDefaultBucket, map[string]any{
			"Name":          rlName,
			"MaxTokens":     maxTokens,
			"TokensPerFill": tokensPerFill,
			"FillInterval":  "1h",
		}, "istio-system")

		url := fmt.Sprintf("%s/ip", baseURL)
		ratelimitasserts.AssertRateLimitResponseHeaders(t, http.MethodGet, url, nil, maxTokens, maxTokens-1)
	})
}
