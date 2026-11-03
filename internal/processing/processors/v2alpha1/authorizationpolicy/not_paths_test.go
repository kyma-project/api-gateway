package authorizationpolicy

import (
	"slices"
	"testing"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
)

func TestGenerateNotPaths(t *testing.T) {
	const (
		get  = gatewayv2alpha1.HttpMethod("GET")
		post = gatewayv2alpha1.HttpMethod("POST")
	)
	rule := func(path string, methods ...gatewayv2alpha1.HttpMethod) gatewayv2alpha1.Rule {
		return gatewayv2alpha1.Rule{Path: path, Methods: methods}
	}

	tests := []struct {
		name    string
		rules   []gatewayv2alpha1.Rule
		current gatewayv2alpha1.Rule
		want    []string
	}{
		{
			name:    "secured / carved out of later noAuth /{**}",
			rules:   []gatewayv2alpha1.Rule{rule("/", get), rule("/{**}", get)},
			current: rule("/{**}", get),
			want:    []string{"/"},
		},
		{
			name:    "secured / carved out of later noAuth /*",
			rules:   []gatewayv2alpha1.Rule{rule("/", get), rule("/*", get)},
			current: rule("/*", get),
			want:    []string{"/"},
		},
		{
			name:    "secured / rule gets no notPaths",
			rules:   []gatewayv2alpha1.Rule{rule("/", get), rule("/{**}", get)},
			current: rule("/", get),
			want:    nil,
		},
		{
			name:    "no carve when methods do not overlap",
			rules:   []gatewayv2alpha1.Rule{rule("/", get), rule("/{**}", post)},
			current: rule("/{**}", post),
			want:    nil,
		},
		{
			name:    "secured /admin carved out of later noAuth /{**}",
			rules:   []gatewayv2alpha1.Rule{rule("/admin", get), rule("/{**}", get)},
			current: rule("/{**}", get),
			want:    []string{"/admin"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := generateNotPaths(tc.rules, tc.current)
			if !slices.Equal(got, tc.want) {
				t.Errorf("generateNotPaths() = %v, want %v", got, tc.want)
			}
		})
	}
}
