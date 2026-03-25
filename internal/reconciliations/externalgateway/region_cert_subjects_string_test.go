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
  - ugw_hyperscaler_region: "aws/eu-central-1"
    btp_region: "eu10"
    iaas:
      provider: "AWS"
      key: "eu-central-1"
    btp_cf_regions:
      - eu10
    ugw_cert_subjects:
      - "C=DE, O=SAP SE, OU=SAP Cloud Platform Clients, OU=8785f86a-5c84-441f-99cb-1b718a3ba7b8, L=ugw, CN=aws/eu-central-1"
`,
			},
			externalRegion: "eu10",
			expectedSubjects: []string{
				"CN=aws/eu-central-1,L=ugw,OU=8785f86a-5c84-441f-99cb-1b718a3ba7b8,OU=SAP Cloud Platform Clients,O=SAP SE,C=DE",
			},
			expectError: false,
		},
		{
			name: "multiple certificates",
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
      - "C=US, O=Example Inc, OU=Clients, OU=example-uuid-1, L=gateway, CN=provider1/region-a"
      - "C=US, O=Example Inc, OU=Clients, OU=example-uuid-2, L=gateway, CN=provider1/region-a"
`,
			},
			externalRegion: "btp-region-1",
			expectedSubjects: []string{
				"CN=provider1/region-a,L=gateway,OU=example-uuid-1,OU=Clients,O=Example Inc,C=US",
				"CN=provider1/region-a,L=gateway,OU=example-uuid-2,OU=Clients,O=Example Inc,C=US",
			},
			expectError: false,
		},
		{
			name: "region not found",
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
      - "C=US, O=Example Inc, L=gateway, CN=provider1/region-a"
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
					BTPRegion:        tt.externalRegion,
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
