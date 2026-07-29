package ipfamily_test

import (
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
)

func TestFromEnv(t *testing.T) {
	cases := []struct {
		env      string
		want     ipfamily.Family
		policy   corev1.IPFamilyPolicy
		families []corev1.IPFamily
		networks []string
	}{
		{"", ipfamily.IPv4Only, corev1.IPFamilyPolicySingleStack, []corev1.IPFamily{corev1.IPv4Protocol}, []string{"tcp4"}},
		{"ipv4", ipfamily.IPv4Only, corev1.IPFamilyPolicySingleStack, []corev1.IPFamily{corev1.IPv4Protocol}, []string{"tcp4"}},
		{"ipv6", ipfamily.IPv6Only, corev1.IPFamilyPolicySingleStack, []corev1.IPFamily{corev1.IPv6Protocol}, []string{"tcp6"}},
		{"dualstack", ipfamily.DualStack, corev1.IPFamilyPolicyPreferDualStack, []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}, []string{"tcp4", "tcp6"}},
	}

	for _, tc := range cases {
		t.Run(tc.env, func(t *testing.T) {
			t.Setenv("TEST_IP_FAMILY", tc.env)
			f := ipfamily.From()
			if f != tc.want {
				t.Fatalf("From() = %q, want %q", f, tc.want)
			}
			if got := f.Policy(); got != tc.policy {
				t.Errorf("Policy() = %q, want %q", got, tc.policy)
			}
			if got := f.Families(); !slices.Equal(got, tc.families) {
				t.Errorf("Families() = %v, want %v", got, tc.families)
			}
			if got := f.DialNetworks(); !slices.Equal(got, tc.networks) {
				t.Errorf("DialNetworks() = %v, want %v", got, tc.networks)
			}
		})
	}
}

func TestFromInvalidPanics(t *testing.T) {
	t.Setenv("TEST_IP_FAMILY", "bogus")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on invalid TEST_IP_FAMILY")
		}
	}()
	_ = ipfamily.From()
}

func TestApplyToService(t *testing.T) {
	// ClusterIP (and empty type, which Kubernetes defaults to ClusterIP)
	// receive the family axis.
	for _, tc := range []struct {
		name         string
		env          string
		svcType      corev1.ServiceType
		wantPolicy   corev1.IPFamilyPolicy
		wantFamilies []corev1.IPFamily
	}{
		{"empty type ipv4 default", "", "", corev1.IPFamilyPolicySingleStack, []corev1.IPFamily{corev1.IPv4Protocol}},
		{"ClusterIP dualstack", "dualstack", corev1.ServiceTypeClusterIP, corev1.IPFamilyPolicyPreferDualStack, []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}},
		{"ClusterIP ipv6", "ipv6", corev1.ServiceTypeClusterIP, corev1.IPFamilyPolicySingleStack, []corev1.IPFamily{corev1.IPv6Protocol}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEST_IP_FAMILY", tc.env)
			svc := &corev1.Service{Spec: corev1.ServiceSpec{Type: tc.svcType}}
			ipfamily.ApplyToService(svc)
			if svc.Spec.IPFamilyPolicy == nil || *svc.Spec.IPFamilyPolicy != tc.wantPolicy {
				t.Errorf("IPFamilyPolicy = %v, want %v", svc.Spec.IPFamilyPolicy, tc.wantPolicy)
			}
			if !slices.Equal(svc.Spec.IPFamilies, tc.wantFamilies) {
				t.Errorf("IPFamilies = %v, want %v", svc.Spec.IPFamilies, tc.wantFamilies)
			}
		})
	}

	// Non-ClusterIP Services pass through unchanged — istio owns the
	// istio-ingressgateway LB and the mutator must not touch it.
	for _, svcType := range []corev1.ServiceType{corev1.ServiceTypeLoadBalancer, corev1.ServiceTypeNodePort, corev1.ServiceTypeExternalName} {
		t.Run("skips "+string(svcType), func(t *testing.T) {
			t.Setenv("TEST_IP_FAMILY", "dualstack")
			svc := &corev1.Service{Spec: corev1.ServiceSpec{Type: svcType}}
			ipfamily.ApplyToService(svc)
			if svc.Spec.IPFamilyPolicy != nil {
				t.Errorf("IPFamilyPolicy should be nil for %s, got %v", svcType, *svc.Spec.IPFamilyPolicy)
			}
			if svc.Spec.IPFamilies != nil {
				t.Errorf("IPFamilies should be nil for %s, got %v", svcType, svc.Spec.IPFamilies)
			}
		})
	}
}
