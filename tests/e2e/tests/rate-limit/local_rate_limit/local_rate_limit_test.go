package local_rate_limit

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	ratelimitasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/ratelimit"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/domain"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/testsetup"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/decoder"
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
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

	kymaGatewayDomain, err := domain.GetFromGateway(t, "kyma-gateway", "kyma-system")
	require.NoError(t, err, "Failed to get domain from kyma-gateway")

	t.Run("Pod rate limited by default bucket", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-default"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitDefaultBucket, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)

		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, nil))
	})

	t.Run("Pod rate limited by default bucket with IPv6 client", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-default-v6"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitDefaultBucket, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		headers := map[string]string{"X-Forwarded-For": "::1"}
		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, headers))
	})

	t.Run("Pod rate limited by path-based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-path"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitPathBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, nil))
	})

	t.Run("Pod rate limited by path-based configuration with IPv6 client", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-path-v6"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitPathBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		headers := map[string]string{"X-Forwarded-For": "::1"}
		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, headers))
	})

	t.Run("Pod rate limited by header-based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-header"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitHeaderBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		headers := map[string]string{"X-Rate-Limited": "true"}
		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, headers))
	})

	t.Run("Pod rate limited by header-based configuration with IPv6 client", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-header-v6"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitHeaderBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/ip", testBackground.TestName, kymaGatewayDomain)
		headers := map[string]string{
			"X-Rate-Limited":  "true",
			"X-Forwarded-For": "::1",
		}
		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, headers))
	})

	t.Run("Pod rate limited by path and header based configuration", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-path-hdr"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitPathAndHeaderBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain)
		headers := map[string]string{"X-Rate-Limited": "true"}
		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, headers))
	})

	t.Run("Pod rate limited by path and header based configuration with IPv6 client", func(t *testing.T) {
		t.Parallel()
		testBackground, err := testsetup.SetupRandomNamespaceWithHttpbin(t, testsetup.WithPrefix("rl-ph-v6"))
		require.NoError(t, err)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitAPIRule, map[string]any{
			"TestID":           testBackground.TestName,
			"Namespace":        testBackground.Namespace,
			"GatewayNamespace": "kyma-system",
			"GatewayName":      "kyma-gateway",
			"Domain":           kymaGatewayDomain,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		apiruleasserts.WaitUntilReady(t, "apirule-"+testBackground.TestName, testBackground.Namespace)

		_, err = infrahelpers.CreateResourceWithTemplateValues(t, RateLimitPathAndHeaderBased, map[string]any{
			"Name":      testBackground.TestName,
			"Namespace": testBackground.Namespace,
		}, decoder.MutateNamespace(testBackground.Namespace))
		require.NoError(t, err)

		ratelimitasserts.WaitUntilReady(t, testBackground.TestName, testBackground.Namespace)

		url := fmt.Sprintf("https://%s.%s/headers", testBackground.TestName, kymaGatewayDomain)
		headers := map[string]string{
			"X-Rate-Limited":  "true",
			"X-Forwarded-For": "::1",
		}
		require.NoError(t, ratelimitasserts.AssertEventuallyRateLimited(t, http.MethodGet, url, headers))
	})

}
