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

	t.Run("Ingress gateway rate limited by default bucket", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-igw-default"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayDefaultBucket, map[string]any{
			"Name": testBackground.TestName,
		}, "istio-system")

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, nil)
	})

	t.Run("Ingress gateway rate limited by path-based configuration", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-igw-path"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathBased, map[string]any{
			"Name": testBackground.TestName,
		}, "istio-system")

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, nil)
	})

	t.Run("Ingress gateway rate limited by header-based configuration", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-igw-hdr"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayHeaderBased, map[string]any{
			"Name": testBackground.TestName,
		}, "istio-system")

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, map[string]string{"X-Rate-Limited": "true"})
	})

	t.Run("Ingress gateway rate limited by path and header based configuration", func(t *testing.T) {
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-igw-ph"))
		require.NoError(t, err, "Failed to setup test namespace with httpbin")

		ratelimitasserts.SetupAPIRule(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, testBackground.Namespace)

		ratelimitasserts.SetupRateLimit(t, RateLimitIngressGatewayPathAndHeaderBased, map[string]any{
			"Name": testBackground.TestName,
		}, "istio-system")

		url := fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain)
		ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, map[string]string{"X-Rate-Limited": "true"})
	})
}
