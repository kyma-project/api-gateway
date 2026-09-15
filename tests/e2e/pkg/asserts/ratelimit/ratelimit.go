package ratelimit

import (
	"context"
	_ "embed"
	"net/http"
	"strconv"
	"strings"
	"testing"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

func WaitUntilReady(t *testing.T, name, namespace string) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err, "Failed to create resources client")

	var rl ratelimitv1alpha1.RateLimit
	require.NoError(t, r.Get(t.Context(), name, namespace, &rl), "Failed to get RateLimit %s/%s", namespace, name)

	err = wait.For(conditions.New(r).ResourceMatch(&rl, func(obj k8s.Object) bool {
		cr, ok := obj.(*ratelimitv1alpha1.RateLimit)
		if !ok {
			t.Fatalf("Expected object of type RateLimit, got %T", obj)
		}
		return cr.Status.State == ratelimitv1alpha1.StatusReady
	}))
	if err != nil {
		t.Logf("RateLimit %s/%s status: %+v", namespace, name, rl.Status)
	}
	require.NoError(t, err)
}

// SetupAPIRule creates an APIRule from the given template and waits until it is ready.
func SetupAPIRule(t *testing.T, apiRuleYAML string, templateValues map[string]any, namespace string) {
	t.Helper()

	_, err := infrahelpers.CreateResourceWithTemplateValues(t, apiRuleYAML, templateValues, decoder.MutateNamespace(namespace))
	require.NoError(t, err, "Failed to create APIRule resource")

	apiruleasserts.WaitUntilReady(t, "apirule-"+templateValues["TestID"].(string), namespace)
}

// SetupRateLimit creates a RateLimit resource from the given template and waits until it is ready.
// It also registers a cleanup that waits for the resource to be fully deleted, so that sequential
// tests sharing the same ingressgateway do not observe stale Envoy state.
func SetupRateLimit(t *testing.T, rateLimitYAML string, templateValues map[string]any, namespace string) {
	t.Helper()

	name := templateValues["Name"].(string)

	// Register WaitUntilDeleted before CreateResourceWithTemplateValues so that in Lifo cleanup
	// order it runs last - after the delete issued by createResource's own cleanup.
	setup.DeclareCleanup(t, func() {
		WaitUntilDeleted(t, name, namespace)
	})

	_, err := infrahelpers.CreateResourceWithTemplateValues(t, rateLimitYAML, templateValues, decoder.MutateNamespace(namespace))
	require.NoError(t, err, "Failed to create RateLimit resource")

	WaitUntilReady(t, name, namespace)
}

func WaitUntilDeleted(t *testing.T, name, namespace string) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err, "Failed to create resources client")

	rl := &ratelimitv1alpha1.RateLimit{}
	rl.SetName(name)
	rl.SetNamespace(namespace)

	err = wait.For(conditions.New(r).ResourceDeleted(rl))
	require.NoError(t, err, "RateLimit %s/%s was not deleted within timeout", namespace, name)
}

// AssertNSuccessfulResponses sends n requests in total and asserts that each
// response is 200 OK. In dualstack mode, requests are distributed per family.
func AssertNSuccessfulResponses(t *testing.T, n int, method, url string, headers map[string]string) {
	t.Helper()
	require.Greater(t, n, 0, "n must be greater than 0")

	networks := ipfamily.From().DialNetworks()
	perFamily := n / len(networks)
	if perFamily == 0 {
		perFamily = 1
	}

	ipfamily.ForEachDialNetwork(t, "rate-limit", nil, func(t *testing.T, _ string, httpClient *http.Client) {
		for i := 0; i < perFamily; i++ {
			req, err := http.NewRequest(method, url, nil)
			require.NoError(t, err, "failed to create request")
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			resp, err := httpClient.Do(req)
			require.NoErrorf(t, err, "request error for %s", url)
			_ = resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				assert.Failf(t, "unexpected status", "expected 200 OK from %s on request %d/%d, got %d", url, i+1, perFamily, resp.StatusCode)
				return
			}
		}
	})
}

// AssertNSuccessfulResponsesWithRetries sends up to maxRetries requests and asserts that at least n
// 200 OK responses are received. 429 responses are ignored. The function logs the total requests sent
// and how many were successful.
//
// Use this instead of AssertNSuccessfulResponses when there are multiple pods:
// each pod has an independent token bucket, so requests must be distributed across all pods to drain
// their combined capacity. maxRetries provides a budget to account for uneven LB distribution.
func AssertNSuccessfulResponsesWithRetries(t *testing.T, n, maxRetries int, method, url string, headers map[string]string) {
	t.Helper()
	require.Greater(t, n, 0, "n must be greater than 0")
	require.GreaterOrEqual(t, maxRetries, n, "maxRetries must be >= n")

	networks := ipfamily.From().DialNetworks()
	perFamily := n / len(networks)
	if perFamily == 0 {
		perFamily = 1
	}
	maxRetriesPerFamily := maxRetries / len(networks)
	if maxRetriesPerFamily == 0 {
		maxRetriesPerFamily = 1
	}

	ipfamily.ForEachDialNetwork(t, "rate-limit", nil, func(t *testing.T, _ string, httpClient *http.Client) {
		successful := 0
		totalRequests := 0

		for totalRequests < maxRetriesPerFamily {
			req, err := http.NewRequest(method, url, nil)
			require.NoError(t, err, "failed to create request")
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			resp, err := httpClient.Do(req)
			require.NoErrorf(t, err, "request error for %s", url)
			_ = resp.Body.Close()
			totalRequests++

			if resp.StatusCode == http.StatusOK {
				successful++
			}

			if successful == perFamily {
				break
			}
		}

		t.Logf("total requests sent: %d, successful: %d", totalRequests, successful)
		assert.GreaterOrEqualf(t, successful, perFamily, "expected at least %d successful responses from %s", perFamily, url)
	})
}

// AssertNRateLimitedResponses sends n requests in total and asserts that each
// response is 429 TooManyRequests. In dualstack mode, requests are distributed per family.
func AssertNRateLimitedResponses(t *testing.T, n int, method, url string, headers map[string]string) {
	t.Helper()
	require.Greater(t, n, 0, "n must be greater than 0")

	networks := ipfamily.From().DialNetworks()
	perFamily := n / len(networks)
	if perFamily == 0 {
		perFamily = 1
	}

	ipfamily.ForEachDialNetwork(t, "rate-limit", nil, func(t *testing.T, _ string, httpClient *http.Client) {
		for i := 0; i < perFamily; i++ {
			req, err := http.NewRequest(method, url, nil)
			require.NoError(t, err, "failed to create request")
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			resp, err := httpClient.Do(req)
			require.NoErrorf(t, err, "request error for %s", url)
			_ = resp.Body.Close()

			if resp.StatusCode != http.StatusTooManyRequests {
				assert.Failf(t, "unexpected status", "expected 429 TooManyRequests from %s on request %d/%d, got %d", url, i+1, perFamily, resp.StatusCode)
				return
			}
		}
	})
}

// AssertRateLimitResponseHeaders asserts that the response contains the x-ratelimit-limit
// and x-ratelimit-remaining headers with the expected values, which are set by Envoy when
// enableResponseHeaders is true.
func AssertRateLimitResponseHeaders(t *testing.T, method, url string, headers map[string]string, expectedLimit, expectedRemaining int) {
	t.Helper()

	ipfamily.ForEachDialNetwork(t, "rate-limit", nil, func(t *testing.T, _ string, httpClient *http.Client) {
		req, err := http.NewRequest(method, url, nil)
		require.NoError(t, err, "failed to create request")
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := httpClient.Do(req)
		require.NoErrorf(t, err, "request error for %s", url)
		_ = resp.Body.Close()

		assert.Equal(t, strconv.Itoa(expectedLimit), resp.Header.Get("x-ratelimit-limit"), "unexpected x-ratelimit-limit header")
		assert.Equal(t, strconv.Itoa(expectedRemaining), resp.Header.Get("x-ratelimit-remaining"), "unexpected x-ratelimit-remaining header")
	})
}

// MatchingPodCount returns the number of running pods in namespace whose labels
// match the selectorLabels of the named RateLimit. Returns 1 if none are found.
func MatchingPodCount(t *testing.T, rateLimitName, namespace string) int {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err, "failed to create resources client")

	var rl ratelimitv1alpha1.RateLimit
	require.NoError(t, r.Get(t.Context(), rateLimitName, namespace, &rl), "failed to get RateLimit %s/%s", namespace, rateLimitName)

	k8sClient, err := client.GetClientSet(t)
	require.NoError(t, err, "failed to create k8s clientset")

	selector := labelMapToSelector(rl.Spec.SelectorLabels)
	pods, err := k8sClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector,
	})
	require.NoError(t, err, "failed to list pods for RateLimit %s/%s", namespace, rateLimitName)

	count := 0
	for _, p := range pods.Items {
		if p.Status.Phase == "Running" {
			count++
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

func labelMapToSelector(labels map[string]string) string {
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}
