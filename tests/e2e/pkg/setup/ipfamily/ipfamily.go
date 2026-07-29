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
	"os"

	corev1 "k8s.io/api/core/v1"
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
