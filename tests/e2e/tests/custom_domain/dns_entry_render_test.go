package custom_domain

import (
	"strings"
	"testing"
	"text/template"
)

// TestDNSEntryTemplateRendersAllTargets verifies the DNSEntry fixture emits one
// target per entry in the Targets slice, using the same text/template engine and
// missingkey=error option as infrahelpers.CreateResourceWithTemplateValues.
func TestDNSEntryTemplateRendersAllTargets(t *testing.T) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(DNSEntryTemplate)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, map[string]any{
		"Name":      "cd-test",
		"Subdomain": "sub.example.com",
		"Targets":   []string{"10.0.0.1", "2001:db8::1"},
	})
	if err != nil {
		t.Fatalf("execute template: %v", err)
	}

	out := sb.String()
	for _, want := range []string{`- "10.0.0.1"`, `- "2001:db8::1"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered manifest missing %q:\n%s", want, out)
		}
	}
}
