package extauth

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/dnswait"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
)

// AssertEndpoint dials `url` and asserts the response status. When
// TEST_IP_FAMILY selects more than one network the request is run once
// per family inside a t.Run(<network>, ...) sub-test; single-family runs
// match pre-dualstack behaviour byte for byte. Failures are reported via
// subtest t.Errorf / t.Fatal — no error return.
func AssertEndpoint(t *testing.T, method, url string, headers map[string]string, expectedHttpCode int) {
	t.Helper()
	// On Gardener AWS the APIRule host resolves only after the DNSEntry
	// propagates externally (~30-90s, occasionally longer). Skipping this
	// leaves the test to dial NXDOMAIN and fail milliseconds after
	// APIRule.Status is Ready.
	dnswait.WaitForURL(t, url)

	ipfamily.ForEachDialNetwork(t, "ext-auth-client", nil, func(t *testing.T, _ string, client *http.Client) {
		request, err := http.NewRequest(method, url, nil)
		require.NoError(t, err, "failed to create request")
		for headerName, headerValue := range headers {
			request.Header.Set(headerName, headerValue)
		}

		response, err := client.Do(request)
		require.NoError(t, err, "failed to perform request")
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(response.Body)
		assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")
	})
}

// AssertEndpointWithJWT dials `url` with a JWT provider and asserts the
// response status. Failures are reported via subtest t.Errorf / t.Fatal
// (through the underlying require/assert helpers) and cause the subtest
// to fail; there is no error return because there is no assertion-level
// error a caller would meaningfully act on.
//
// The JWT provider owns its HTTP client construction, so network-family
// pinning at this layer is not currently available; when TEST_IP_FAMILY
// is dualstack the request runs once per family under t.Run(<network>,
// ...) but the underlying HTTP client is unpinned. On k3d
// (TEST_IP_FAMILY unset) behaviour is byte-identical to before.
func AssertEndpointWithJWT(t *testing.T, method, url string, expectedHttpCode int, provider oauth2.Provider, options ...oauth2.RequestOption) {
	t.Helper()
	// Wait for external DNS on both the APIRule host (dialled below via
	// provider.MakeRequest) and, transitively, the provider's TokenURL
	// (same wildcard shoot domain — one wait covers both since
	// WaitForURL keys off the host and the mock lives under the same
	// wildcard as the APIRule).
	dnswait.WaitForURL(t, url)

	// Ignore the client — the JWT provider owns its own HTTP client. We
	// still use ForEachDialNetwork to get the per-family t.Run wrapping in
	// dualstack mode; single-family mode runs once inline.
	ipfamily.ForEachDialNetwork(t, "ext-auth-jwt", nil, func(t *testing.T, _ string, _ *http.Client) {
		statusCode, _, _, err := provider.MakeRequest(t, method, url, options...)
		require.NoError(t, err, "failed to make request with JWT")
		assert.Equal(t, expectedHttpCode, statusCode, "unexpected status code")
	})
}
