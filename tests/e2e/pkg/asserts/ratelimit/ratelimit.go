package ratelimit

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	apiruleasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/apirule"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NoError(t, err)
	if err != nil {
		t.Logf("RateLimit %s/%s status: %+v", namespace, name, rl.Status)
	}
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
	t.Cleanup(func() {
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
	assert.NoError(t, err, "RateLimit %s/%s was not deleted within timeout", namespace, name)
}

// AssertEventuallyRateLimited repeatedly sends requests until a 429 is received,
// failing the test if no 429 is returned within the deadline. When TEST_IP_FAMILY
// selects more than one network (dualstack), the assertion runs once per family.
func AssertEventuallyRateLimited(t *testing.T, method, url string, headers map[string]string) {
	t.Helper()

	ipfamily.ForEachDialNetwork(t, "rate-limit", nil, func(t *testing.T, _ string, httpClient *http.Client) {
		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			req, err := http.NewRequest(method, url, nil)
			if err != nil {
				assert.NoError(t, fmt.Errorf("failed to create request: %w", err))
				return
			}
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				t.Logf("request error: %v — retrying", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
		assert.Fail(t, fmt.Sprintf("expected 429 TooManyRequests from %s but did not receive it within deadline", url))
	})
}

// AssertNotRateLimited sends requestCount requests and fails immediately if any
// returns 429. When TEST_IP_FAMILY selects more than one network (dualstack),
// the assertion runs once per family.
func AssertNotRateLimited(t *testing.T, method, url string, headers map[string]string, requestCount int) {
	t.Helper()

	ipfamily.ForEachDialNetwork(t, "rate-limit", nil, func(t *testing.T, _ string, httpClient *http.Client) {
		for i := 0; i < requestCount; i++ {
			req, err := http.NewRequest(method, url, nil)
			if err != nil {
				assert.NoError(t, fmt.Errorf("failed to create request: %w", err))
				return
			}
			for k, v := range headers {
				req.Header.Set(k, v)
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				t.Logf("request error: %v", err)
				continue
			}
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				assert.Fail(t, fmt.Sprintf("unexpected 429 TooManyRequests from %s on request %d/%d", url, i+1, requestCount))
				return
			}
		}
	})
}
