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

var externalRegionsConfigMapName = "external-gateway-regions"

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
regions:
- ugw_hyperscaler_region: "provider1/region-a"
  btp_region: btp-region-1
  iaas:
    provider: Provider1
    key: region-a
  btp_cf_regions:
    - btp-region-1
    - btp-region-1-001
  ugw_cert_subjects:
    - "C=US, O=Example Inc, OU=Clients, OU=example-uuid-1, L=gateway, CN=provider1/region-a"
    - "C=US, O=Example Inc, OU=Clients, OU=example-uuid-2, L=gateway, CN=provider1/region-a"
  ugw_inbound_static_ips:
    - 10.0.1.100
    - 10.0.1.101
  ugw_outbound_static_ips:
    - 10.0.2.100
    - 10.0.2.101
- ugw_hyperscaler_region: provider2/region-b
  btp_region: btp-region-2
  iaas:
    provider: Provider2
    key: region-b
  btp_cf_regions:
    - btp-region-2
  ugw_cert_subjects:
    - "C=US, O=Example Inc, OU=Clients, OU=example-uuid-1, L=gateway, CN=provider2/region-b"
`,
			},
			externalRegion: "btp-region-1",
			expectedSubjects: []RegionCertSubject{
				{
					CN: "provider1/region-a",
					C:  "US",
					O:  "Example Inc",
					L:  "gateway",
					OU: []string{"Clients", "example-uuid-1"},
				},
				{
					CN: "provider1/region-a",
					C:  "US",
					O:  "Example Inc",
					L:  "gateway",
					OU: []string{"Clients", "example-uuid-2"},
				},
			},
			expectError: false,
		},
		{
			name: "single key with different name (regions_examples)",
			configMapData: map[string]string{
				"regions_examples": `
regions:
  - ugw_hyperscaler_region: "cloudprovider/region-east-1"
    btp_region: "btp-east-10"
    iaas:
      provider: "CloudProvider"
      key: "region-east-1"
    btp_cf_regions:
      - btp-east-10
    ugw_cert_subjects:
      - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=cloudprovider/region-east-1"
    ugw_inbound_static_ips: []
    ugw_outbound_static_ips: []
`,
			},
			externalRegion: "btp-east-10",
			expectedSubjects: []RegionCertSubject{
				{
					CN: "cloudprovider/region-east-1",
					C:  "US",
					O:  "Example Inc",
					L:  "gateway",
					OU: []string{"Clients"},
				},
			},
			expectError: false,
		},
		{
			name: "case-insensitive region matching",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - ugw_hyperscaler_region: "PROVIDER1/REGION-A"
    btp_region: "BTP-REGION-1"
    iaas:
      provider: "PROVIDER1"
      key: "REGION-A"
    btp_cf_regions:
      - BTP-REGION-1
    ugw_cert_subjects:
      - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=provider1/region-a"
    ugw_inbound_static_ips: []
    ugw_outbound_static_ips: []
`,
			},
			externalRegion: "btp-region-1",
			expectedSubjects: []RegionCertSubject{
				{
					CN: "provider1/region-a",
					C:  "US",
					O:  "Example Inc",
					L:  "gateway",
					OU: []string{"Clients"},
				},
			},
			expectError: false,
		},
		{
			name: "region not found in ConfigMap",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
- ugw_hyperscaler_region: "provider1/region-a"
  btp_region: "btp-region-1"
  iaas:
    provider: "Provider1"
    key: "region-a"
  btp_cf_regions:
    - btp-region-1
  ugw_cert_subjects:
    - "C=US, O=Example Inc, OU=Clients, L=gateway, CN=provider1/region-a"
`,
			},
			externalRegion:   "btp-nonexistent",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "not found in ConfigMap",
		},
		{
			name: "multiple regions specified - only first processed",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - ugw_hyperscaler_region: "provider1/region-a"
    btp_region: "btp-region-1"
    iaas:
      provider: "Provider1"
      key: "region-a"
    btp_cf_regions:
      - btp-region-1
    ugw_cert_subjects:
      - "C=US, O=Example Inc, OU=shared-ou, OU=region-specific-1, L=gateway, CN=provider1/region-a"
    ugw_inbound_static_ips: []
    ugw_outbound_static_ips: []
  - ugw_hyperscaler_region: "provider1/region-b"
    btp_region: "btp-region-2"
    iaas:
      provider: "Provider1"
      key: "region-b"
    btp_cf_regions:
      - btp-region-2
    ugw_cert_subjects:
      - "C=US, O=Example Inc, OU=shared-ou, OU=region-specific-2, L=gateway, CN=provider1/region-b"
    ugw_inbound_static_ips: []
    ugw_outbound_static_ips: []
`,
			},
			externalRegion: "btp-region-1",
			expectedSubjects: []RegionCertSubject{
				{
					CN: "provider1/region-a",
					C:  "US",
					O:  "Example Inc",
					L:  "gateway",
					OU: []string{"shared-ou", "region-specific-1"},
				},
			},
			expectError: false,
		},
		{
			name:             "ConfigMap missing regions.yaml key",
			configMapData:    map[string]string{},
			externalRegion:   "btp-east-10",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "is empty",
		},
		{
			name: "invalid YAML format",
			configMapData: map[string]string{
				"regions.yaml": `invalid: yaml: [[[`,
			},
			externalRegion:   "btp-east-10",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "failed to parse regions.yaml",
		},
		{
			name: "empty regions list in YAML",
			configMapData: map[string]string{
				"regions.yaml": `regions: []`,
			},
			externalRegion:   "btp-east-10",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "no regions found in ConfigMap",
		},
		{
			name: "region with empty cert subjects",
			configMapData: map[string]string{
				"regions.yaml": `
regions:
  - ugw_hyperscaler_region: "cloudprovider/region-east-1"
    btp_region: "btp-east-10"
    iaas:
      provider: "CloudProvider"
      key: "region-east-1"
    btp_cf_regions:
      - btp-east-10
    ugw_cert_subjects: []
    ugw_inbound_static_ips: []
    ugw_outbound_static_ips: []
`,
			},
			externalRegion:   "btp-east-10",
			expectedSubjects: nil,
			expectError:      true,
			errorContains:    "no certificate subjects found for requested region: btp-east-10",
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
					BTPRegion:        tt.externalRegion,
					RegionsConfigMap: externalRegionsConfigMapName,
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
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
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
					if actual.CN != expected.CN {
						t.Errorf("subject[%d].CN: expected '%s', got '%s'", i, expected.CN, actual.CN)
					}
					if actual.C != expected.C {
						t.Errorf("subject[%d].C: expected '%s', got '%s'", i, expected.C, actual.C)
					}
					if actual.O != expected.O {
						t.Errorf("subject[%d].O: expected '%s', got '%s'", i, expected.O, actual.O)
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
			BTPRegion:        "btp-east-10",
			RegionsConfigMap: externalRegionsConfigMapName,
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
	if !strings.Contains(err.Error(), "failed to get ConfigMap") {
		t.Errorf("error message '%s' should mention ConfigMap not found", err.Error())
	}
}
