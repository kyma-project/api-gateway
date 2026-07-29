package endpoint

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
// Failures are reported via subtest t.Errorf / t.Fatal (through the
// underlying require/assert helpers). There is no error return: callers
// have no assertion-level error to act on.
func AssertEndpoint(t *testing.T, method, url string, expectedHttpCode int) {
	t.Helper()
	dnswait.WaitForURL(t, url)
	ipfamily.ForEachDialNetwork(t, "http-client-go", nil, func(t *testing.T, _ string, client *http.Client) {
		doAssert(t, client, method, url, nil, expectedHttpCode, nil, nil)
	})
}

func AssertEndpointWithoutResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedMissingHeaders []string) {
	t.Helper()
	dnswait.WaitForURL(t, url)
	ipfamily.ForEachDialNetwork(t, "http-client-go", nil, func(t *testing.T, _ string, client *http.Client) {
		doAssert(t, client, method, url, requestHeaders, expectedHttpCode, nil, expectedMissingHeaders)
	})
}

func AssertEndpointWithResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string) {
	t.Helper()
	dnswait.WaitForURL(t, url)
	ipfamily.ForEachDialNetwork(t, "http-client-go", nil, func(t *testing.T, _ string, client *http.Client) {
		doAssert(t, client, method, url, requestHeaders, expectedHttpCode, expectedResponseHeaders, nil)
	})
}

// doAssert issues the request, closes the body, and asserts status +
// header expectations. Pre-response failures (NewRequest, Do) are reported
// via require.NoError on the subtest so the failure attaches to the right
// test node and halts the subtest.
func doAssert(t *testing.T, httpClient *http.Client, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string, expectedMissingHeaders []string) {
	t.Helper()
	request, err := http.NewRequest(method, url, nil)
	require.NoError(t, err, "failed to create request")
	for headerName, headerValue := range requestHeaders {
		request.Header.Set(headerName, headerValue)
	}

	response, err := httpClient.Do(request)
	require.NoError(t, err, "failed to perform request")
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
}
