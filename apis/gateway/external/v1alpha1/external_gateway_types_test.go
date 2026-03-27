/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBaseName(t *testing.T) {
	tests := []struct {
		name           string
		inputName      string
		expectedMaxLen int // max length of BaseName
		expectTruncate bool
	}{
		{
			name:           "short name unchanged",
			inputName:      "my-app",
			expectedMaxLen: 45,
			expectTruncate: false,
		},
		{
			name:           "name at limit unchanged",
			inputName:      strings.Repeat("a", 45),
			expectedMaxLen: 45,
			expectTruncate: false,
		},
		{
			name:           "name over limit - truncated with hash",
			inputName:      strings.Repeat("a", 60),
			expectedMaxLen: 45,
			expectTruncate: true,
		},
		{
			name:           "very long name - truncated with hash",
			inputName:      "my-very-long-application-name-for-external-gateway-resource-that-exceeds-limit",
			expectedMaxLen: 45,
			expectTruncate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ExternalGateway{
				ObjectMeta: metav1.ObjectMeta{Name: tt.inputName},
			}
			result := e.BaseName()

			// Verify length constraint
			if len(result) > tt.expectedMaxLen {
				t.Errorf("BaseName too long: got %d chars, expected max %d", len(result), tt.expectedMaxLen)
			}

			// Verify truncation behavior
			if tt.expectTruncate {
				if result == tt.inputName {
					t.Errorf("expected name to be truncated but got original: %s", result)
				}
				// Should contain hash suffix (7 hex chars after dash)
				if !strings.Contains(result, "-") {
					t.Errorf("truncated name should contain dash for hash suffix: %s", result)
				}
			} else {
				if result != tt.inputName {
					t.Errorf("expected name unchanged, got: %s, want: %s", result, tt.inputName)
				}
			}

			// Verify deterministic (same input = same output)
			if e.BaseName() != result {
				t.Error("BaseName not deterministic")
			}

			// Verify all derived names are under 63 chars
			derivedNames := map[string]string{
				"GatewayName":              e.GatewayName(),
				"CertificateName":          e.CertificateName(),
				"TLSSecretName":            e.TLSSecretName(),
				"CASecretName":             e.CASecretName(),
				"DNSEntryName":             e.DNSEntryName(),
				"XFCCFilterName":           e.XFCCFilterName(),
				"CertValidationFilterName": e.CertValidationFilterName(),
			}

			for methodName, derivedName := range derivedNames {
				if len(derivedName) > 63 {
					t.Errorf("%s exceeds 63 chars: got %d chars (%s)", methodName, len(derivedName), derivedName)
				}
			}
		})
	}
}

func TestDerivedNamesConsistency(t *testing.T) {
	e := &ExternalGateway{
		ObjectMeta: metav1.ObjectMeta{Name: "test-gateway"},
	}

	// Verify suffixes
	if e.GatewayName() != "test-gateway-gw" {
		t.Errorf("GatewayName: got %s, want test-gateway-gw", e.GatewayName())
	}
	if e.CertificateName() != "test-gateway-cert" {
		t.Errorf("CertificateName: got %s, want test-gateway-cert", e.CertificateName())
	}
	if e.TLSSecretName() != "test-gateway-tls" {
		t.Errorf("TLSSecretName: got %s, want test-gateway-tls", e.TLSSecretName())
	}
	if e.CASecretName() != "test-gateway-tls-cacert" {
		t.Errorf("CASecretName: got %s, want test-gateway-tls-cacert", e.CASecretName())
	}
	if e.DNSEntryName() != "test-gateway-dns" {
		t.Errorf("DNSEntryName: got %s, want test-gateway-dns", e.DNSEntryName())
	}
	if e.XFCCFilterName() != "test-gateway-xfcc" {
		t.Errorf("XFCCFilterName: got %s, want test-gateway-xfcc", e.XFCCFilterName())
	}
	if e.CertValidationFilterName() != "test-gateway-cv" {
		t.Errorf("CertValidationFilterName: got %s, want test-gateway-cv", e.CertValidationFilterName())
	}
}
