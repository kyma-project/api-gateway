package endpoint

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/dnswait"
	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
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
	return runPerFamilyOnURL(t, "http-client-go", url, func(t *testing.T, client *http.Client) error {
		return doAssert(t, client, method, url, nil, expectedHttpCode, nil, nil)
	})
}

func AssertEndpointWithoutResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedMissingHeaders []string) error {
	t.Helper()
	return runPerFamilyOnURL(t, "http-client-go", url, func(t *testing.T, client *http.Client) error {
		return doAssert(t, client, method, url, requestHeaders, expectedHttpCode, nil, expectedMissingHeaders)
	})
}

func AssertEndpointWithResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string) error {
	t.Helper()
	return runPerFamilyOnURL(t, "http-client-go", url, func(t *testing.T, client *http.Client) error {
		return doAssert(t, client, method, url, requestHeaders, expectedHttpCode, expectedResponseHeaders, nil)
	})
}

// runPerFamily invokes fn once per configured dial network. When only one
// family is configured (default, or ipv4/ipv6 single-family mode) it runs
// fn directly on t so gotestsum output stays byte-identical to today.
// DualStack mode opens a t.Run(<network>, ...) sub-test per family with an
// HTTP client whose transport is pinned to that TCP family.
//
// fn returns an error only for pre-response failures (build request, dial
// error at the transport). Assertion mismatches use t.Errorf inside fn.
// The returned error is the last non-nil fn return; callers currently pass
// it through require.NoError so the first pre-response failure fails the
// test. Multi-family runs surface per-family failures through subtest
// reporting independently of this return value.
func runPerFamily(t *testing.T, prefix string, fn func(t *testing.T, client *http.Client) error) error {
	t.Helper()
	return runPerFamilyOnURL(t, prefix, "", fn)
}

// runPerFamilyOnURL is like runPerFamily but also waits for external DNS to
// publish records for `dialURL`'s host per family before dispatching. On
// Gardener AWS, DNSEntry propagation lags the in-cluster Ready state by
// 30-90s per family; without this wait, tests dial NXDOMAIN and fail
// instantly. Skipped for empty URL, host-less URLs, IP literals, and
// hostnames without a dot (localhost, kubernetes.default, etc.) so k3d
// and in-cluster paths pay no cost.
func runPerFamilyOnURL(t *testing.T, prefix string, dialURL string, fn func(t *testing.T, client *http.Client) error) error {
	t.Helper()
	networks := ipfamily.From().DialNetworks()

	host := hostFromURL(dialURL)

	if len(networks) == 1 {
		// Single-family fast path. Match the pre-dualstack behaviour byte
		// for byte on k3d: no sub-test wrapping, no per-family log prefix.
		network := networks[0]
		if needsDNSWait(host) {
			waitDNSForFamily(t, host, network)
		}
		client := httphelper.NewHTTPClient(t,
			httphelper.WithPrefix(prefix),
			httphelper.WithNetwork(networkForClient(network)),
		)
		return fn(t, client)
	}

	var lastErr error
	for _, network := range networks {
		t.Run(prefix+"-"+network, func(t *testing.T) {
			if needsDNSWait(host) {
				waitDNSForFamily(t, host, network)
			}
			client := httphelper.NewHTTPClient(t,
				httphelper.WithPrefix(prefix+"-"+network),
				httphelper.WithNetwork(network),
			)
			if err := fn(t, client); err != nil {
				lastErr = err
				t.Errorf("request failed for %s: %v", network, err)
			}
		})
	}
	return lastErr
}

// hostFromURL parses `raw` and returns its hostname (no port). Empty
// string when parsing fails or no host is present, which callers treat
// as "skip the DNS wait".
func hostFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// needsDNSWait skips the wait for IP literals and single-label hosts
// (localhost, service names). Only dotted DNS names potentially served by
// an external resolver actually need the pre-dial stability check.
func needsDNSWait(host string) bool {
	if host == "" {
		return false
	}
	if net.ParseIP(host) != nil {
		return false
	}
	return strings.Contains(host, ".")
}

// waitDNSForFamily blocks until at least one address of `network`'s
// family (tcp4 / tcp6) resolves stably for `host`. On failure it logs
// with t.Logf and returns — the subsequent HTTP dial will surface the
// real error rather than a wait-timeout wrapper, keeping the test
// output focused on the actual assertion failure.
func waitDNSForFamily(t *testing.T, host, network string) {
	t.Helper()
	// Slightly longer than dnswait.Timeout so the inner poll fires its
	// own timeout error (which names the missing family) before this
	// context deadline chops the call off with a bare
	// context.DeadlineExceeded.
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	if err := dnswait.WaitForHost(ctx, host, []string{network}); err != nil {
		t.Logf("DNS wait for %q family %s did not stabilise: %v (proceeding to dial anyway)", host, network, err)
	}
}

// networkForClient returns the WithNetwork value for a given DialNetworks
// entry. On the single-family fast path we want the client to keep the
// resolver-default behaviour on IPv4 clusters (where TEST_IP_FAMILY is
// unset), so an "ipv4" family passes an empty network — matching the
// pre-dualstack behaviour where NewHTTPClient did not override DialContext
// at all. Any explicit v6-only mode still pins the family.
func networkForClient(network string) string {
	if ipfamily.From() == ipfamily.IPv4Only {
		// Preserve the pre-dualstack transport (no custom DialContext).
		return ""
	}
	return network
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
