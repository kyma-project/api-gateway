package extauth

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/dnswait"
	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
)

// AssertEndpoint dials `url` and asserts the response status. When
// TEST_IP_FAMILY selects more than one network, the request is run once
// per family in a t.Run(<network>, ...) sub-test. Callers see a single
// call; single-family runs match pre-dualstack behaviour byte for byte.
func AssertEndpoint(t *testing.T, method, url string, headers map[string]string, expectedHttpCode int) error {
	t.Helper()
	// Same DNS-wait as endpoint.AssertEndpoint uses. On Gardener AWS the
	// APIRule host resolves only after the DNSEntry propagates externally
	// (~30-90 s, occasionally longer). Skipping this leaves the test to
	// dial NXDOMAIN and fail milliseconds after APIRule.Status is Ready.
	dnswait.WaitForURL(t, url)
	networks := ipfamily.From().DialNetworks()

	run := func(t *testing.T, client *http.Client) error {
		request, err := http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		for headerName, headerValue := range headers {
			request.Header.Set(headerName, headerValue)
		}

		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("failed to perform request: %w", err)
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(response.Body)
		assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")
		return nil
	}

	if len(networks) == 1 {
		network := ""
		if ipfamily.From() != ipfamily.IPv4Only {
			network = networks[0]
		}
		return run(t, httphelper.NewHTTPClient(t,
			httphelper.WithPrefix("ext-auth-client"),
			httphelper.WithNetwork(network),
		))
	}

	var lastErr error
	for _, network := range networks {
		t.Run(network, func(t *testing.T) {
			client := httphelper.NewHTTPClient(t,
				httphelper.WithPrefix("ext-auth-client-"+network),
				httphelper.WithNetwork(network),
			)
			if err := run(t, client); err != nil {
				lastErr = err
				t.Errorf("request failed for %s: %v", network, err)
			}
		})
	}
	return lastErr
}

// AssertEndpointWithJWT dials `url` with a JWT provider and asserts the
// response status. The JWT provider owns its HTTP client construction, so
// network-family pinning at this layer is not currently available; when
// TEST_IP_FAMILY is dualstack the request runs once per family under
// t.Run(<network>, ...) but the underlying HTTP client is unpinned. On k3d
// (TEST_IP_FAMILY unset) behaviour is byte-identical to before.
func AssertEndpointWithJWT(t *testing.T, method, url string, expectedHttpCode int, provider oauth2.Provider, options ...oauth2.RequestOption) error {
	t.Helper()
	// Wait for external DNS on both the APIRule host (dialled below via
	// provider.MakeRequest) and, transitively, the provider's TokenURL
	// (same wildcard shoot domain — one wait covers both since
	// WaitForURL keys off the host and the mock lives under the same
	// wildcard as the APIRule).
	dnswait.WaitForURL(t, url)
	networks := ipfamily.From().DialNetworks()

	run := func(t *testing.T) error {
		statusCode, _, _, err := provider.MakeRequest(t, method, url, options...)
		require.NoError(t, err, "failed to make request with JWT")
		assert.Equal(t, expectedHttpCode, statusCode, "unexpected status code")
		return nil
	}

	if len(networks) == 1 {
		return run(t)
	}

	var lastErr error
	for _, network := range networks {
		t.Run(network, func(t *testing.T) {
			if err := run(t); err != nil {
				lastErr = err
			}
		})
	}
	return lastErr
}
