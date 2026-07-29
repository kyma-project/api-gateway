package dnswait

import (
	"net"
	"slices"
	"sort"
	"testing"
)

func TestHostFromURL(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"invalid", "://bad", ""},
		{"https with port", "https://api.example.com:443/foo", "api.example.com"},
		{"http no port", "http://api.example.com/foo", "api.example.com"},
		{"ip literal", "http://127.0.0.1/", "127.0.0.1"},
		{"ipv6 literal", "http://[::1]:8080/", "::1"},
		{"path only", "/foo/bar", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hostFromURL(tc.raw); got != tc.want {
				t.Errorf("hostFromURL(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestNeedsWait(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"", false},
		{"localhost", false},
		{"kubernetes.default", true}, // dotted; caller decides to skip via other means
		{"api.example.com", true},
		{"127.0.0.1", false},
		{"::1", false},
		{"httpbin.local.kyma.dev", true},
	}
	for _, tc := range cases {
		t.Run(tc.host, func(t *testing.T) {
			if got := needsWait(tc.host); got != tc.want {
				t.Errorf("needsWait(%q) = %v, want %v", tc.host, got, tc.want)
			}
		})
	}
}

func TestNormaliseAddrs(t *testing.T) {
	// Two lookups returning the same set in different order compare equal.
	a := []net.IP{net.ParseIP("10.0.0.2"), net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.3")}
	b := []net.IP{net.ParseIP("10.0.0.3"), net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2")}
	got := normaliseAddrs(a)
	want := normaliseAddrs(b)
	if !slices.Equal(got, want) {
		t.Errorf("normaliseAddrs is not order-insensitive: %v vs %v", got, want)
	}
	if !sort.StringsAreSorted(got) {
		t.Errorf("normaliseAddrs did not sort output: %v", got)
	}
}

func TestSlicesEqual(t *testing.T) {
	cases := []struct {
		a, b []string
		want bool
	}{
		{nil, nil, true},
		{[]string{}, []string{}, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a", "b"}, []string{"a"}, false},
		{[]string{"a", "b"}, []string{"b", "a"}, false}, // order matters (input is expected pre-sorted)
	}
	for _, tc := range cases {
		if got := slicesEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("slicesEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
