// Package dnswait blocks a test until an external DNS resolver has
// published a stable record set for a hostname on each requested address
// family. Motivation: on Gardener AWS, the api-gateway module creates a
// DNSEntry whose external propagation (Route 53 -> public resolvers) lags
// the in-cluster Ready state by 30-90 seconds, occasionally longer when
// the LoadBalancer NLB has just been recreated. Dialling in that window
// hits NXDOMAIN and the test fails milliseconds after ApiGateway.Status
// says everything is ready. This helper closes that window without
// coupling the assertion path to controller-runtime clients or Service
// status — the only signal it needs is the URL the test is about to
// dial.
package dnswait

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
)

// Timeout bounds how long WaitForHost waits for a single family's record
// set to stabilise. Six minutes reflects the observed worst case on
// Gardener AWS when back-to-back test suites tear down and recreate the
// APIGateway CR: each recreate cycles the NLB, forcing a fresh Route 53
// publication and negative-cache invalidation upstream. Typical
// propagation is 30-90 s; the extra headroom absorbs the tail.
const Timeout = 6 * time.Minute

// WaitForHost polls the host resolver until every dial network in
// `networks` (values from ipfamily.DialNetworks, e.g. "tcp4"/"tcp6")
// returns a STABLE non-empty address set. Stability means two
// consecutive polls returned the same set of addresses (order-
// insensitive). This closes the window between "APIGateway.Status is
// Ready" and "public DNS has published records for both families",
// and also protects against partial propagation: an AWS NLB in a
// 3-AZ shoot registers up to 3 A records per family, and Route 53
// can publish them staggered — dialling after the first record
// appears may hit an ENI whose data-plane is not yet ready. Requiring
// stability (matching consecutive results) avoids dialling into the
// transient middle state. On timeout the error names the family that
// never stabilised, so callers see e.g. `hostname X: no ip4 addresses`
// instead of a generic per-request NXDOMAIN retry loop. Parent-context
// cancellation is surfaced unwrapped so callers can errors.Is-check it.
func WaitForHost(t *testing.T, ctx context.Context, host string, networks []string) error {
	t.Helper()
	// Resolver "ip4" / "ip6" mirror the socket-family filter Go's dialer
	// applies for "tcp4" / "tcp6".
	ipNetworkFor := map[string]string{"tcp4": "ip4", "tcp6": "ip6"}

	for _, n := range networks {
		ipNet, ok := ipNetworkFor[n]
		if !ok {
			return fmt.Errorf("dnswait: unsupported network %q (want tcp4 or tcp6)", n)
		}
		lastErr := fmt.Errorf("no lookup attempted")
		attempt := 0
		var previous []string
		err := wait.PollUntilContextTimeout(ctx, 5*time.Second, Timeout, true, func(ctx context.Context) (bool, error) {
			attempt++
			addrs, err := net.DefaultResolver.LookupIP(ctx, ipNet, host)
			if err != nil {
				lastErr = err
				t.Logf("dnswait: %s lookup for %q attempt %d failed: %v", ipNet, host, attempt, err)
				previous = nil
				return false, nil
			}
			if len(addrs) == 0 {
				lastErr = fmt.Errorf("no %s addresses", ipNet)
				t.Logf("dnswait: %s lookup for %q attempt %d returned no addresses", ipNet, host, attempt)
				previous = nil
				return false, nil
			}
			current := normaliseAddrs(addrs)
			t.Logf("dnswait: %s lookup for %q attempt %d returned %d addr(s): %v", ipNet, host, attempt, len(current), current)
			if previous != nil && slices.Equal(previous, current) {
				return true, nil
			}
			// Either first successful poll or set changed — record and
			// wait another interval to confirm stability.
			lastErr = fmt.Errorf("%s address set not yet stable: %v", ipNet, current)
			previous = current
			return false, nil
		})
		if err != nil {
			// Preserve parent-context cancellation / deadline as-is so
			// callers can errors.Is against context.Canceled or
			// context.DeadlineExceeded. Only wrap when the poll's own
			// timeout fired.
			if ctxErr := ctx.Err(); ctxErr != nil && errors.Is(err, ctxErr) {
				return err
			}
			return fmt.Errorf("hostname %q: no stable %s addresses after %s: %w", host, ipNet, Timeout, lastErr)
		}
	}
	return nil
}

// normaliseAddrs converts a slice of net.IP into a sorted slice of
// strings so two lookups returning the same set in different order
// compare equal.
func normaliseAddrs(addrs []net.IP) []string {
	out := make([]string, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, a.String())
	}
	sort.Strings(out)
	return out
}

// urlWaitTimeout is the outer bound for WaitForURL's context — slightly
// longer than the inner per-family Timeout so the poll fires its own
// family-named error before the context deadline chops the call off with
// a bare context.DeadlineExceeded.
const urlWaitTimeout = Timeout + time.Minute

// WaitForURL parses `dialURL`, extracts the hostname, and blocks until the
// caller's configured dial networks (via ipfamily.From().DialNetworks())
// resolve stably. Skipped for empty URL, host-less URLs, IP literals, and
// single-label hosts (localhost, kubernetes.default, etc.) so k3d and
// in-cluster paths pay no cost. Failures are logged via t.Logf; the
// subsequent dial by the caller will surface the real error rather than
// a wait-timeout wrapper, keeping the test output focused on the actual
// assertion failure. This is the entry point for callers outside the
// endpoint fan-out (e.g. the oauth2 JWT provider path) that need the
// same pre-dial DNS guarantee.
func WaitForURL(t *testing.T, dialURL string) {
	t.Helper()
	host := hostFromURL(dialURL)
	if !needsWait(host) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), urlWaitTimeout)
	defer cancel()
	if err := WaitForHost(t, ctx, host, ipfamily.From().DialNetworks()); err != nil {
		t.Logf("DNS wait for %q did not stabilise: %v (proceeding to dial anyway)", host, err)
	}
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

// needsWait skips the wait for IP literals and single-label hosts
// (localhost, service names). Only dotted DNS names potentially served by
// an external resolver actually need the pre-dial stability check.
func needsWait(host string) bool {
	if host == "" {
		return false
	}
	if net.ParseIP(host) != nil {
		return false
	}
	return strings.Contains(host, ".")
}
