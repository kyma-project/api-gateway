package ingress

import (
	"net"
	"strings"
	"testing"
)

func TestLoadBalancerTargets(t *testing.T) {
	tests := []struct {
		name    string
		ingress []any
		want    []string
		wantErr bool
	}{
		{
			name: "dual-stack per-family IPs",
			ingress: []any{
				map[string]any{"ip": "10.0.0.1"},
				map[string]any{"ip": "2001:db8::1"},
			},
			want: []string{"10.0.0.1", "2001:db8::1"},
		},
		{
			name:    "single hostname",
			ingress: []any{map[string]any{"hostname": "lb.example.com"}},
			want:    []string{"lb.example.com"},
		},
		{
			name:    "single IP",
			ingress: []any{map[string]any{"ip": "10.0.0.1"}},
			want:    []string{"10.0.0.1"},
		},
		{
			name:    "empty ingress",
			ingress: []any{},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loadBalancerTargets(tc.ingress)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalStrings(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIPFamiliesFromSpec(t *testing.T) {
	tests := []struct {
		name          string
		spec          map[string]any
		want          []string
		wantDualStack bool
	}{
		{
			name:          "dual-stack",
			spec:          specWithFamilies("IPv4", "IPv6"),
			want:          []string{"IPv4", "IPv6"},
			wantDualStack: true,
		},
		{
			name: "single IPv4",
			spec: specWithFamilies("IPv4"),
			want: []string{"IPv4"},
		},
		{
			name: "single IPv6",
			spec: specWithFamilies("IPv6"),
			want: []string{"IPv6"},
		},
		{
			name: "empty defaults to IPv4",
			spec: map[string]any{"spec": map[string]any{}},
			want: []string{"IPv4"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ipFamiliesFromSpec(tc.spec)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalStrings(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if lb := (LoadBalancer{IPFamilies: got}); lb.DualStack() != tc.wantDualStack {
				t.Fatalf("DualStack(%v) = %v, want %v", got, lb.DualStack(), tc.wantDualStack)
			}
		})
	}
}

func specWithFamilies(families ...string) map[string]any {
	f := make([]any, len(families))
	for i, v := range families {
		f[i] = v
	}
	return map[string]any{"spec": map[string]any{"ipFamilies": f}}
}

// TestProbeMatch drives the unified readiness core. requireBoth is applied by the caller
// (WaitUntilDNSReady): dual-stack needs v4 && v6, single-stack needs v4 || v6.
func TestProbeMatch(t *testing.T) {
	restore := lookupIP
	t.Cleanup(func() { lookupIP = restore })

	tests := []struct {
		name          string
		probe         []net.IP
		targetLookup  map[string][]net.IP
		targets       []string
		wantV4        bool
		wantV6        bool
	}{
		{
			name:    "both families present matching IP targets",
			probe:   []net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("2001:db8::1")},
			targets: []string{"10.0.0.1", "2001:db8::1"},
			wantV4:  true,
			wantV6:  true,
		},
		{
			name:    "only IPv4 resolves",
			probe:   []net.IP{net.ParseIP("10.0.0.1")},
			targets: []string{"10.0.0.1", "2001:db8::1"},
			wantV4:  true,
			wantV6:  false,
		},
		{
			name:         "hostname target resolving to both families (AWS NLB CNAME)",
			probe:        []net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("2001:db8::1")},
			targetLookup: map[string][]net.IP{"lb.example.com": {net.ParseIP("10.0.0.1"), net.ParseIP("2001:db8::1")}},
			targets:      []string{"lb.example.com"},
			wantV4:       true,
			wantV6:       true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookupIP = func(host string) ([]net.IP, error) {
				if strings.HasPrefix(host, "probe-") {
					return tc.probe, nil
				}
				if ips, ok := tc.targetLookup[host]; ok {
					return ips, nil
				}
				return nil, &net.DNSError{Err: "no such host", Name: host}
			}
			v4, v6, err := probeMatch("sub.example.com", tc.targets)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v4 != tc.wantV4 || v6 != tc.wantV6 {
				t.Fatalf("probeMatch = (v4=%v, v6=%v), want (v4=%v, v6=%v)", v4, v6, tc.wantV4, tc.wantV6)
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
