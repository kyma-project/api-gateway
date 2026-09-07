package ratelimit

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

func WaitUntilReady(t *testing.T, name, namespace string) {
	t.Helper()

	r, err := client.ResourcesClient(t)
	require.NoError(t, err)

	var rl ratelimitv1alpha1.RateLimit
	require.NoError(t, r.Get(t.Context(), name, namespace, &rl))

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

// AssertEventuallyRateLimited repeatedly sends requests until a 429 is received,
// failing the test if no 429 is returned within the deadline.
func AssertEventuallyRateLimited(t *testing.T, method, url string, headers map[string]string) error {
	t.Helper()

	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("rate-limit"))

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
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

		t.Logf("response: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("expected 429 TooManyRequests from %s but did not receive it within deadline", url)
}
