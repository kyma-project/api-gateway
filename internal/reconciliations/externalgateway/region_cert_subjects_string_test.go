package externalgateway

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

func TestResolveRegionCertSubjects_StringOutput(t *testing.T) {
	tests := []struct {
		name             string
		configMapData    map[string]string
		externalRegion   string
		expectedSubjects []string
		expectError      bool
		errorContains    string
	}{
		{
			name: "single certificate - reversed order",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - name: "eu10"
    ips:
      - 10.0.0.1
    subjects:
      - "C=DE, O=Example Corp, OU=Cloud Platform Clients, OU=a1b2c3d4-e5f6-7890-abcd-ef1234567890, L=gateway, CN=aws/eu-central-1"
`,
			},
			externalRegion: "eu10",
			expectedSubjects: []string{
				"CN=aws/eu-central-1,L=gateway,OU=a1b2c3d4-e5f6-7890-abcd-ef1234567890,OU=Cloud Platform Clients,O=Example Corp,C=DE",
			},
			expectError: false,
		},
		{
			name: "multiple certificates",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - name: "us10"
    ips:
      - 10.0.0.2
    subjects:
      - "C=US, O=Example Inc, OU=Clients, OU=uuid-1111-2222-3333-444444444444, L=gateway, CN=aws/us-east-1"
      - "C=US, O=Example Inc, OU=Clients, OU=uuid-5555-6666-7777-888888888888, L=gateway, CN=aws/us-east-1"
`,
			},
			externalRegion: "us10",
			expectedSubjects: []string{
				"CN=aws/us-east-1,L=gateway,OU=uuid-1111-2222-3333-444444444444,OU=Clients,O=Example Inc,C=US",
				"CN=aws/us-east-1,L=gateway,OU=uuid-5555-6666-7777-888888888888,OU=Clients,O=Example Inc,C=US",
			},
			expectError: false,
		},
		{
			name: "region not found",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - name: "eu10"
    ips:
      - 10.0.0.1
    subjects:
      - "C=US, O=Example Inc, L=gateway, CN=aws/eu-central-1"
`,
			},
			externalRegion:   "non-existent-region",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			external := &externalv1alpha1.ExternalGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
				Spec: externalv1alpha1.ExternalGatewaySpec{
					Region:           tt.externalRegion,
					RegionsConfigMap: "external-gateway-regions",
				},
			}

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "external-gateway-regions",
					Namespace: "test-namespace",
				},
				Data: tt.configMapData,
			}

			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)
			_ = externalv1alpha1.AddToScheme(scheme)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(configMap).
				Build()

			// Execute
			ctx := context.Background()
			subjects, err := ResolveRegionCertSubjects(ctx, fakeClient, external)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error message '%s' does not contain '%s'", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Check subjects (string comparison)
			if !tt.expectError {
				if len(subjects) != len(tt.expectedSubjects) {
					t.Errorf("expected %d subjects, got %d", len(tt.expectedSubjects), len(subjects))
				}
				for i, expected := range tt.expectedSubjects {
					if i >= len(subjects) {
						t.Errorf("missing subject at index %d", i)
						break
					}
					actual := subjects[i]
					if actual != expected {
						t.Errorf("subject[%d] mismatch:\nexpected: %s\ngot:      %s", i, expected, actual)
					}
				}
			}
		})
	}
}
