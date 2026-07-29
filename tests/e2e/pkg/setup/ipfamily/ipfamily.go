// Package ipfamily selects Kubernetes IP family behaviour for e2e test
// fixtures and clients based on the TEST_IP_FAMILY environment variable.
//
// TEST_IP_FAMILY values: "ipv4" (default), "ipv6", "dualstack". Any other
// value panics — the intent is to fail loudly in CI rather than silently
// drift back to the default.
//
// The fan-out over families lives inside assertion helpers (see
// tests/e2e/pkg/asserts/...), so test files stay one-liners and cannot
// forget to iterate the configured families. Callers who need the raw
// axis use From().DialNetworks().
package ipfamily

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"

	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
)

type Family string

const (
	IPv4Only  Family = "ipv4"
	IPv6Only  Family = "ipv6"
	DualStack Family = "dualstack"

	EnvVar = "TEST_IP_FAMILY"
)

// From reads TEST_IP_FAMILY and returns the selected Family. Empty defaults
// to IPv4Only; anything unrecognised panics.
func From() Family {
	v := os.Getenv(EnvVar)
	switch v {
	case "", string(IPv4Only):
		return IPv4Only
	case string(IPv6Only):
		return IPv6Only
	case string(DualStack):
		return DualStack
	default:
		panic(fmt.Sprintf("ipfamily: unrecognised %s=%q (want ipv4|ipv6|dualstack)", EnvVar, v))
	}
}

// Policy returns the Kubernetes Service.spec.ipFamilyPolicy for this family.
func (f Family) Policy() corev1.IPFamilyPolicy {
	switch f {
	case DualStack:
		return corev1.IPFamilyPolicyPreferDualStack
	default:
		return corev1.IPFamilyPolicySingleStack
	}
}

// Families returns the ordered list to set on Service.spec.ipFamilies.
// DualStack lists IPv6 first so pods and Services report a v6 primary
// address; tests that read podIPs[0] observe the v6 side.
func (f Family) Families() []corev1.IPFamily {
	switch f {
	case IPv4Only:
		return []corev1.IPFamily{corev1.IPv4Protocol}
	case IPv6Only:
		return []corev1.IPFamily{corev1.IPv6Protocol}
	case DualStack:
		return []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}
	}
	return nil
}

// DialNetworks returns the Go `net` network strings a test should exercise
// for this family. IPv4Only / IPv6Only return a single-element slice pinning
// that family; DualStack returns both so a test asserts the service works
// over v4 AND v6. Pass a returned value as the `network` argument to
// `net.Dialer.DialContext`, or use "tcp4"/"tcp6" with an `http.Transport`
// custom DialContext.
func (f Family) DialNetworks() []string {
	switch f {
	case IPv4Only:
		return []string{"tcp4"}
	case IPv6Only:
		return []string{"tcp6"}
	case DualStack:
		return []string{"tcp4", "tcp6"}
	}
	return nil
}

// ApplyToService writes ipFamilyPolicy + ipFamilies onto svc based on the
// current TEST_IP_FAMILY. Skipped for Service types that don't accept the
// fields (ExternalName has no pod-side IPs). Idempotent — safe to call
// from a decoder.MutateOption on every fixture. Under TEST_IP_FAMILY=""
// or "ipv4" the resulting spec matches the pre-dualstack SingleStack IPv4
// default
func ApplyToService(svc *corev1.Service) {
	if svc.Spec.Type != "" && svc.Spec.Type != corev1.ServiceTypeClusterIP {
		return
	}
	f := From()
	policy := f.Policy()
	svc.Spec.IPFamilyPolicy = &policy
	svc.Spec.IPFamilies = f.Families()
}

// ForEachDialNetwork runs fn once per dial network configured by
// TEST_IP_FAMILY. Single-family runs (TEST_IP_FAMILY unset or ipv4/ipv6)
// invoke fn once inline with no t.Run wrapping — matching pre-dualstack
// behaviour byte for byte on k3d, where TEST_IP_FAMILY is unset and no
// custom DialContext should be installed. DualStack invocations open one
// t.Run(<network>, ...) sub-test per family, each with an http.Client
// pinned to that TCP family via httphelper.WithNetwork.
//
// The caller-supplied `prefix` is used as the httphelper log prefix; in
// dualstack mode it is suffixed with "-<network>" so per-family log lines
// are self-labelling. Options in `opts` are appended after the invariant
// WithPrefix/WithNetwork so a caller can override them if they legitimately
// need to.
//
// This is the canonical way to exercise LB dials against dualstack shoots
// from an assertion helper. Callers that need external DNS to stabilise
// before dialling should call dnswait.WaitForURL(t, url) themselves before
// invoking this helper — that mirrors the pattern already used by the
// oauth2 mock and keeps the fan-out shape single-purpose.
func ForEachDialNetwork(t *testing.T, prefix string, opts []httphelper.Option, fn func(t *testing.T, network string, client *http.Client) error) error {
	t.Helper()
	networks := From().DialNetworks()

	if len(networks) == 1 {
		// Preserve pre-dualstack behaviour on k3d: TEST_IP_FAMILY unset =>
		// IPv4Only => no custom DialContext, client keeps resolver-default
		// dialling. An explicit v6-only mode still pins the family.
		network := ""
		if From() != IPv4Only {
			network = networks[0]
		}
		all := append([]httphelper.Option{
			httphelper.WithPrefix(prefix),
			httphelper.WithNetwork(network),
		}, opts...)
		return fn(t, networks[0], httphelper.NewHTTPClient(t, all...))
	}

	var lastErr error
	for _, network := range networks {
		t.Run(network, func(t *testing.T) {
			all := append([]httphelper.Option{
				httphelper.WithPrefix(prefix + "-" + network),
				httphelper.WithNetwork(network),
			}, opts...)
			if err := fn(t, network, httphelper.NewHTTPClient(t, all...)); err != nil {
				lastErr = err
				t.Errorf("request failed for %s: %v", network, err)
			}
		})
	}
	return lastErr
}
