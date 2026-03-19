package externalgateway

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

func TestResolveCertSubjects(t *testing.T) {
	tests := []struct {
		name             string
		configMapData    map[string]string
		externalRegion   string
		expectedSubjects []RegionCertSubject
		expectError      bool
		errorContains    string
	}{
		{
			name: "multiple regions specified - only first region processed",
			configMapData: map[string]string{
				"regions.yaml": `
- Provider: provider1
  Region: region-a
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, OU=uuid-1, L=gateway, CN=provider1/region-a"
    - "C=US, O=Example Inc, OU=Clients, OU=uuid-2, L=gateway, CN=provider1/region-a"
- Provider: provider2
  Region: region-b
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, OU=uuid-1, L=gateway, CN=provider2/region-b"
`,
			},
			externalRegion: "provider1/region-a",
			expectedSubjects: []RegionCertSubject{
				{
					Region: "provider1/region-a",
					CN:     "provider1/region-a",
					L:      "gateway",
					OU:     []string{"Clients", "uuid-1"},
				},
				{
					Region: "provider1/region-a",
					CN:     "provider1/region-a",
					L:      "gateway",
					OU:     []string{"Clients", "uuid-2"},
				},
			},
			expectError: false,
		},
		{
			name: "single key with different name (regions_examples)",
			configMapData: map[string]string{
				"regions_examples": `
- Provider: aws
  Region: us-east-1
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=aws/us-east-1"
`,
			},
			externalRegion: "aws/us-east-1",
			expectedSubjects: []RegionCertSubject{
				{
					Region: "aws/us-east-1",
					CN:     "aws/us-east-1",
					L:      "gateway",
					OU:     []string{"Clients"},
				},
			},
			expectError: false,
		},
		{
			name: "case-insensitive region matching",
			configMapData: map[string]string{
				"regions.yaml": `
- Provider: PROVIDER1
  Region: REGION-A
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=provider1/region-a"
`,
			},
			externalRegion: "provider1/region-a",
			expectedSubjects: []RegionCertSubject{
				{
					Region: "provider1/region-a",
					CN:     "provider1/region-a",
					L:      "gateway",
					OU:     []string{"Clients"},
				},
			},
			expectError: false,
		},
		{
			name: "mixed case in ExternalGateway spec",
			configMapData: map[string]string{
				"regions.yaml": `
- Provider: provider1
  Region: region-a
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=provider1/region-a"
`,
			},
			externalRegion: "PROVIDER1/REGION-A",
			expectedSubjects: []RegionCertSubject{
				{
					Region: "provider1/region-a",
					CN:     "provider1/region-a",
					L:      "gateway",
					OU:     []string{"Clients"},
				},
			},
			expectError: false,
		},
		{
			name: "region not found in ConfigMap",
			configMapData: map[string]string{
				"regions.yaml": `
- Provider: provider1
  Region: region-a
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=provider1/region-a"
`,
			},
			externalRegion:   "provider2/region-b",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "not found in ConfigMap",
		},
		{
			name: "partial match - one region found, one not found - only first used",
			configMapData: map[string]string{
				"regions.yaml": `
- Provider: provider1
  Region: region-a
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=provider1/region-a"
`,
			},
			externalRegion: "provider1/region-a",
			expectedSubjects: []RegionCertSubject{
				{
					Region: "provider1/region-a",
					CN:     "provider1/region-a",
					L:      "gateway",
					OU:     []string{"Clients"},
				},
			},
			expectError: false,
		},
		{
			name: "multiple regions specified - only first processed",
			configMapData: map[string]string{
				"regions.yaml": `
- Provider: provider1
  Region: region-a
  CertSubjects:
    - "C=US, O=Example Inc, OU=shared-ou, OU=region-specific-1, L=gateway, CN=provider1/region-a"
- Provider: provider1
  Region: region-b
  CertSubjects:
    - "C=US, O=Example Inc, OU=shared-ou, OU=region-specific-2, L=gateway, CN=provider1/region-b"
`,
			},
			externalRegion: "provider1/region-a",
			expectedSubjects: []RegionCertSubject{
				{
					Region: "provider1/region-a",
					CN:     "provider1/region-a",
					L:      "gateway",
					OU:     []string{"shared-ou", "region-specific-1"},
				},
			},
			expectError: false,
		},
		{
			name:             "ConfigMap missing regions.yaml key",
			configMapData:    map[string]string{},
			externalRegion:   "aws/us-east-1",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "is empty",
		},
		{
			name: "invalid YAML format",
			configMapData: map[string]string{
				"regions.yaml": `invalid: yaml: [[[`,
			},
			externalRegion:   "aws/us-east-1",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "failed to parse regions.yaml",
		},
		{
			name: "empty regions list in YAML",
			configMapData: map[string]string{
				"regions.yaml": `[]`,
			},
			externalRegion:   "aws/us-east-1",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "not found in ConfigMap",
		},
		{
			name: "region with empty cert subjects",
			configMapData: map[string]string{
				"regions.yaml": `
- Provider: aws
  Region: us-east-1
  CertSubjects: []
`,
			},
			externalRegion:   "aws/us-east-1",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "no certificate subjects found for requested region: aws/us-east-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test objects
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      externalRegionsConfigMapName,
					Namespace: "test-namespace",
				},
				Data: tt.configMapData,
			}

			external := &externalv1alpha1.ExternalGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
				Spec: externalv1alpha1.ExternalGatewaySpec{
					Region: tt.externalRegion,
				},
			}

			// Create fake client
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
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error message '%s' does not contain '%s'", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Check subjects
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
					if actual.Region != expected.Region {
						t.Errorf("subject[%d].Region: expected '%s', got '%s'", i, expected.Region, actual.Region)
					}
					if actual.CN != expected.CN {
						t.Errorf("subject[%d].CN: expected '%s', got '%s'", i, expected.CN, actual.CN)
					}
					if actual.L != expected.L {
						t.Errorf("subject[%d].L: expected '%s', got '%s'", i, expected.L, actual.L)
					}
					if len(actual.OU) != len(expected.OU) {
						t.Errorf("subject[%d].OU: expected %d OUs, got %d", i, len(expected.OU), len(actual.OU))
					} else {
						for j, expectedOU := range expected.OU {
							if actual.OU[j] != expectedOU {
								t.Errorf("subject[%d].OU[%d]: expected '%s', got '%s'", i, j, expectedOU, actual.OU[j])
							}
						}
					}
				}
			}
		})
	}
}

func TestResolveCertSubjects_ConfigMapNotFound(t *testing.T) {
	// Create ExternalGateway without ConfigMap
	external := &externalv1alpha1.ExternalGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gateway",
			Namespace: "test-namespace",
		},
		Spec: externalv1alpha1.ExternalGatewaySpec{
			Region: "aws/us-east-1",
		},
	}

	// Create fake client without ConfigMap
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = externalv1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Execute
	ctx := context.Background()
	_, err := ResolveRegionCertSubjects(ctx, fakeClient, external)

	// Assert
	if err == nil {
		t.Errorf("expected error when ConfigMap not found")
	}
	if !contains(err.Error(), "failed to get ConfigMap") {
		t.Errorf("error message '%s' should mention ConfigMap not found", err.Error())
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}
