package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/dnswait"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
)

// AssertEndpoint dials `url` with the given method and asserts the response
// status matches `expectedHttpCode`. When TEST_IP_FAMILY selects more than
// one network (dualstack), the assertion runs once per family inside a
// t.Run(<network>, ...) sub-test with an HTTP client pinned to that TCP
// family. Callers see a single call; the fan-out is a helper implementation
// detail. Single-family runs stay identical to today (no sub-test wrapping).
//
// The returned error is non-nil only when the request could not be built or
// dispatched at all (before any HTTP response); status-code mismatches are
// reported via t.Errorf inside the subtest, matching testify assertions
// used elsewhere in the codebase.
func AssertEndpoint(t *testing.T, method, url string, expectedHttpCode int) error {
	t.Helper()
	dnswait.WaitForURL(t, url)
	return ipfamily.ForEachDialNetwork(t, "http-client-go", nil, func(t *testing.T, _ string, client *http.Client) error {
		return doAssert(t, client, method, url, nil, expectedHttpCode, nil, nil)
	})
}

func AssertEndpointWithoutResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedMissingHeaders []string) error {
	t.Helper()
	dnswait.WaitForURL(t, url)
	return ipfamily.ForEachDialNetwork(t, "http-client-go", nil, func(t *testing.T, _ string, client *http.Client) error {
		return doAssert(t, client, method, url, requestHeaders, expectedHttpCode, nil, expectedMissingHeaders)
	})
}

func AssertEndpointWithResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string) error {
	t.Helper()
	dnswait.WaitForURL(t, url)
	return ipfamily.ForEachDialNetwork(t, "http-client-go", nil, func(t *testing.T, _ string, client *http.Client) error {
		return doAssert(t, client, method, url, requestHeaders, expectedHttpCode, expectedResponseHeaders, nil)
	})
}

// doAssert issues the request, closes the body, and asserts status +
// header expectations. Returns an error only for pre-response failures.
func doAssert(t *testing.T, httpClient *http.Client, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string, expectedMissingHeaders []string) error {
	t.Helper()
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for headerName, headerValue := range requestHeaders {
		request.Header.Set(headerName, headerValue)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")

	for _, header := range expectedMissingHeaders {
		assert.Empty(t, response.Header.Get(header))
	}
	for headerName, headerValue := range expectedResponseHeaders {
		responseHeaderValue := response.Header.Get(headerName)
		if headerValue != responseHeaderValue {
			t.Fatalf("Didn't get the expected response header: %s: %s, got %s", headerName, headerValue, responseHeaderValue)
		}
	}
	return nil
}
