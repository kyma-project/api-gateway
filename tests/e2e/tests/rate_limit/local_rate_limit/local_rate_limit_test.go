package local_rate_limit

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"
	"time"

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

		baseURL := fmt.Sprintf("https://%s.%s", testBackground.TestName, kymaGatewayDomain)
		rateLimitedPath := "/ip"

		maxTokens := 4
		tokensPerFill := 4

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitDefaultBucket, map[string]any{
			"Name":          testBackground.TestName,
			"Namespace":     testBackground.Namespace,
			"MaxTokens":     maxTokens,
			"TokensPerFill": tokensPerFill,
			"FillInterval":  "1h",
		}, testBackground.Namespace)

		url := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)

		// Assert 'maxTokens' successful responses based on the number of tokens in the default bucket.
		ratelimitasserts.AssertNSuccessfulResponses(t, maxTokens, http.MethodGet, url, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)
	})

	t.Run("Pod rate limited by path-based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-path"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		baseURL := fmt.Sprintf("https://%s.%s", testBackground.TestName, kymaGatewayDomain)
		rateLimitedPath := "/ip"
		nonRateLimitedPath := "/headers"

		defaultMaxTokens := 8
		defaultTokensPerFill := 8
		pathMaxTokens := 4
		pathTokensPerFill := 4

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitPathBased, map[string]any{
			"Name":              testBackground.TestName,
			"Namespace":         testBackground.Namespace,
			"MaxTokens":         defaultMaxTokens,
			"TokensPerFill":     defaultTokensPerFill,
			"FillInterval":      "1h",
			"Path":              rateLimitedPath,
			"PathMaxTokens":     pathMaxTokens,
			"PathTokensPerFill": pathTokensPerFill,
			"PathFillInterval":  "1h",
		}, testBackground.Namespace)

		urlRateLimited := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		urlDefaultBucket := fmt.Sprintf("%s%s", baseURL, nonRateLimitedPath)

		// Assert path-specific bucket limits traffic on the configured path.
		ratelimitasserts.AssertNSuccessfulResponses(t, pathMaxTokens, http.MethodGet, urlRateLimited, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlRateLimited, nil)

		// Assert default bucket limits traffic on paths without path-specific rules.
		ratelimitasserts.AssertNSuccessfulResponses(t, defaultMaxTokens, http.MethodGet, urlDefaultBucket, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlDefaultBucket, nil)
	})

	t.Run("Pod rate limited by header-based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-header"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		baseURL := fmt.Sprintf("https://%s.%s", testBackground.TestName, kymaGatewayDomain)
		rateLimitedPath := "/ip"
		rateLimitedHeaders := map[string]string{"X-Rate-Limited": "true"}
		nonRateLimitedHeaders := map[string]string{"Different-Header": "true"}

		defaultMaxTokens := 8
		defaultTokensPerFill := 8
		headerMaxTokens := 4
		headerTokensPerFill := 4

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitHeaderBased, map[string]any{
			"Name":                testBackground.TestName,
			"Namespace":           testBackground.Namespace,
			"MaxTokens":           defaultMaxTokens,
			"TokensPerFill":       defaultTokensPerFill,
			"FillInterval":        "1h",
			"Header":              "X-Rate-Limited",
			"HeaderValue":         "true",
			"HeaderMaxTokens":     headerMaxTokens,
			"HeaderTokensPerFill": headerTokensPerFill,
			"HeaderFillInterval":  "1h",
		}, testBackground.Namespace)

		url := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		// Assert header-specific bucket applies when the configured header is present.
		ratelimitasserts.AssertNSuccessfulResponses(t, headerMaxTokens, http.MethodGet, url, rateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, rateLimitedHeaders)

		// Assert requests without the configured header are governed by the default bucket.
		ratelimitasserts.AssertNSuccessfulResponses(t, defaultMaxTokens, http.MethodGet, url, nonRateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nonRateLimitedHeaders)
	})

	t.Run("Pod rate limited by path and header based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-path-hdr"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		baseURL := fmt.Sprintf("https://%s.%s", testBackground.TestName, kymaGatewayDomain)
		rateLimitedPath := "/headers"
		nonRateLimitedPath := "/ip"
		rateLimitedHeaders := map[string]string{"X-Rate-Limited": "true"}
		nonRateLimitedHeaders := map[string]string{"Different-Header": "true"}

		defaultMaxTokens := 12
		defaultTokensPerFill := 12
		specializedMaxTokens := 4
		specializedTokensPerFill := 4

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitPathAndHeaderBased, map[string]any{
			"Name":                testBackground.TestName,
			"Namespace":           testBackground.Namespace,
			"MaxTokens":           defaultMaxTokens,
			"TokensPerFill":       defaultTokensPerFill,
			"FillInterval":        "1h",
			"Path":                rateLimitedPath,
			"Header":              "X-Rate-Limited",
			"HeaderValue":         "true",
			"HeaderMaxTokens":     specializedMaxTokens,
			"HeaderTokensPerFill": specializedTokensPerFill,
			"HeaderFillInterval":  "1h",
		}, testBackground.Namespace)

		urlRateLimited := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		urlDefaultBucket := fmt.Sprintf("%s%s", baseURL, nonRateLimitedPath)

		// Assert combined path+header matching is rate limited by the specialized bucket.
		ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlRateLimited, rateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlRateLimited, rateLimitedHeaders)

		// Assert all non-matching combinations consume tokens from the default bucket.
		ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlDefaultBucket, rateLimitedHeaders)
		ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlRateLimited, nonRateLimitedHeaders)
		ratelimitasserts.AssertNSuccessfulResponses(t, specializedMaxTokens, http.MethodGet, urlDefaultBucket, nonRateLimitedHeaders)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, urlDefaultBucket, nonRateLimitedHeaders)
	})

	t.Run("Pod default bucket refills linearly", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-fill"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		baseURL := fmt.Sprintf("https://%s.%s", testBackground.TestName, kymaGatewayDomain)
		rateLimitedPath := "/ip"

		maxTokens := 8
		tokensPerFill := 4
		fillInterval := 8 * time.Second

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitDefaultBucket, map[string]any{
			"Name":          testBackground.TestName,
			"Namespace":     testBackground.Namespace,
			"MaxTokens":     maxTokens,
			"TokensPerFill": tokensPerFill,
			"FillInterval":  fillInterval.String(),
		}, testBackground.Namespace)

		url := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		// Assert the initial tokens are consumed.
		ratelimitasserts.AssertNSuccessfulResponses(t, maxTokens, http.MethodGet, url, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)

		// After half a fill interval (4s), expect tokensPerFill/2 tokens to have been added.
		time.Sleep(4 * time.Second)
		expected := min(int(4*time.Second*time.Duration(tokensPerFill)/fillInterval), maxTokens)
		ratelimitasserts.AssertNSuccessfulResponses(t, expected, http.MethodGet, url, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)

		// After one fill interval (8s), expect tokensPerFill tokens to have been added.
		time.Sleep(8 * time.Second)
		expected = min(int(8*time.Second*time.Duration(tokensPerFill)/fillInterval), maxTokens)
		ratelimitasserts.AssertNSuccessfulResponses(t, expected, http.MethodGet, url, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)

		// After one and a half fill intervals (12s), expect tokensPerFill*3/2 tokens to have been added.
		time.Sleep(12 * time.Second)
		expected = min(int(12*time.Second*time.Duration(tokensPerFill)/fillInterval), maxTokens)
		ratelimitasserts.AssertNSuccessfulResponses(t, expected, http.MethodGet, url, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)
	})

	t.Run("Pod default bucket does not refill past the max", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-fill-max"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		baseURL := fmt.Sprintf("https://%s.%s", testBackground.TestName, kymaGatewayDomain)
		rateLimitedPath := "/ip"

		maxTokens := 2
		tokensPerFill := 2

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitDefaultBucket, map[string]any{
			"Name":          testBackground.TestName,
			"Namespace":     testBackground.Namespace,
			"MaxTokens":     maxTokens,
			"TokensPerFill": tokensPerFill,
			"FillInterval":  "4s",
		}, testBackground.Namespace)

		url := fmt.Sprintf("%s%s", baseURL, rateLimitedPath)
		// Assert the initial tokens are consumed.
		ratelimitasserts.AssertNSuccessfulResponses(t, maxTokens, http.MethodGet, url, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)

		// Wait across multiple intervals to verify refill never exceeds MaxTokens.
		time.Sleep(10 * time.Second)

		// Assert only MaxTokens requests succeed after refill accumulation.
		ratelimitasserts.AssertNSuccessfulResponses(t, maxTokens, http.MethodGet, url, nil)
		ratelimitasserts.AssertNRateLimitedResponses(t, 1, http.MethodGet, url, nil)

	})

	t.Run("Pod response headers present when enableResponseHeaders is true", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-resp-hdr"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		maxTokens := 4
		tokensPerFill := 4

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitDefaultBucket, map[string]any{
			"Name":          testBackground.TestName,
			"Namespace":     testBackground.Namespace,
			"MaxTokens":     maxTokens,
			"TokensPerFill": tokensPerFill,
			"FillInterval":  "1h",
		}, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertRateLimitResponseHeaders(t, http.MethodGet, url, nil, maxTokens, maxTokens-1)
	})

}
