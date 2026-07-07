package external

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations/externalgateway"
)

func TestBuildInternalDomain(t *testing.T) {
	tests := []struct {
		name            string
		subdomain       string
		shootInfoData   map[string]string
		shootInfoExists bool
		expectedDomain  string
		expectError     bool
	}{
		{
			name:      "Gardener available with valid domain",
			subdomain: "my-gateway",
			shootInfoData: map[string]string{
				"domain": "cluster.example.com",
			},
			shootInfoExists: true,
			expectedDomain:  "my-gateway.cluster.example.com",
			expectError:     false,
		},
		{
			name:            "Gardener shoot-info not found - fallback to local.kyma.dev",
			subdomain:       "my-gateway",
			shootInfoData:   nil,
			shootInfoExists: false,
			expectedDomain:  "my-gateway.local.kyma.dev",
			expectError:     false,
		},
		{
			name:      "Gardener shoot-info exists but domain key missing",
			subdomain: "my-gateway",
			shootInfoData: map[string]string{
				"other-key": "other-value",
			},
			shootInfoExists: true,
			expectedDomain:  "",
			expectError:     true,
		},
		{
			name:      "Gardener shoot-info exists but domain empty string",
			subdomain: "my-gateway",
			shootInfoData: map[string]string{
				"domain": "",
			},
			shootInfoExists: true,
			expectedDomain:  "",
			expectError:     true,
		},
		{
			name:            "subdomain with special characters",
			subdomain:       "my-gateway-123",
			shootInfoExists: false,
			expectedDomain:  "my-gateway-123.local.kyma.dev",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create scheme
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)

			// Create fake client
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)

			// Add shoot-info ConfigMap if it should exist
			if tt.shootInfoExists {
				shootInfo := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "shoot-info",
						Namespace: "kube-system",
					},
					Data: tt.shootInfoData,
				}
				clientBuilder = clientBuilder.WithObjects(shootInfo)
			}

			fakeClient := clientBuilder.Build()

			// Create reconciler
			reconciler := &ExternalGatewayReconciler{
				Client: fakeClient,
			}

			// Execute
			ctx := context.Background()
			domain, err := reconciler.buildInternalDomain(ctx, tt.subdomain)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if domain != tt.expectedDomain {
					t.Errorf("expected domain '%s', got '%s'", tt.expectedDomain, domain)
				}
			}
		})
	}
}

func TestGetGardenerDomain(t *testing.T) {
	tests := []struct {
		name            string
		shootInfoData   map[string]string
		shootInfoExists bool
		expectedDomain  string
		expectError     bool
		errorIsNotFound bool
	}{
		{
			name: "valid shoot-info with domain",
			shootInfoData: map[string]string{
				"domain": "cluster.example.com",
			},
			shootInfoExists: true,
			expectedDomain:  "cluster.example.com",
			expectError:     false,
		},
		{
			name:            "shoot-info not found",
			shootInfoExists: false,
			expectError:     true,
			errorIsNotFound: true,
		},
		{
			name: "shoot-info exists but domain key missing",
			shootInfoData: map[string]string{
				"other-key": "value",
			},
			shootInfoExists: true,
			expectError:     true,
			errorIsNotFound: false,
		},
		{
			name: "shoot-info exists but domain empty",
			shootInfoData: map[string]string{
				"domain": "",
			},
			shootInfoExists: true,
			expectError:     true,
			errorIsNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create scheme
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)

			// Create fake client
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)

			if tt.shootInfoExists {
				shootInfo := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "shoot-info",
						Namespace: "kube-system",
					},
					Data: tt.shootInfoData,
				}
				clientBuilder = clientBuilder.WithObjects(shootInfo)
			}

			fakeClient := clientBuilder.Build()

			// Execute
			ctx := context.Background()
			domain, err := getGardenerDomain(ctx, fakeClient)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.errorIsNotFound && !apierrors.IsNotFound(err) {
					t.Errorf("expected NotFound error, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if domain != tt.expectedDomain {
					t.Errorf("expected domain '%s', got '%s'", tt.expectedDomain, domain)
				}
			}
		})
	}
}

func TestGetGardenerShootInfo(t *testing.T) {
	tests := []struct {
		name            string
		shootInfoExists bool
		shootInfoData   map[string]string
		expectError     bool
		errorIsNotFound bool
	}{
		{
			name:            "shoot-info exists",
			shootInfoExists: true,
			shootInfoData: map[string]string{
				"domain": "cluster.example.com",
			},
			expectError: false,
		},
		{
			name:            "shoot-info not found",
			shootInfoExists: false,
			expectError:     true,
			errorIsNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create scheme
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)

			// Create fake client
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)

			if tt.shootInfoExists {
				shootInfo := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "shoot-info",
						Namespace: "kube-system",
					},
					Data: tt.shootInfoData,
				}
				clientBuilder = clientBuilder.WithObjects(shootInfo)
			}

			fakeClient := clientBuilder.Build()

			// Execute
			ctx := context.Background()
			cm, err := getGardenerShootInfo(ctx, fakeClient)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.errorIsNotFound && !apierrors.IsNotFound(err) {
					t.Errorf("expected NotFound error, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if cm == nil {
					t.Errorf("expected ConfigMap but got nil")
				}
				if cm != nil && cm.Name != "shoot-info" {
					t.Errorf("expected ConfigMap name 'shoot-info', got '%s'", cm.Name)
				}
			}
		})
	}
}

func TestBuildInternalDomain_GardenerAPIError(t *testing.T) {
	// This test simulates a non-NotFound error from Gardener API
	// Unfortunately, fake client doesn't easily support simulating API errors
	// So we test the logic path by verifying behavior with missing domain key

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// Create shoot-info without domain key to trigger error
	shootInfo := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shoot-info",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"other-key": "value",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(shootInfo).
		Build()

	reconciler := &ExternalGatewayReconciler{
		Client: fakeClient,
	}

	ctx := context.Background()
	_, err := reconciler.buildInternalDomain(ctx, "test-subdomain")

	// Should return error because domain key is missing
	if err == nil {
		t.Errorf("expected error when domain key missing")
	}

	// Verify it's not a NotFound error (different error path)
	if apierrors.IsNotFound(err) {
		t.Errorf("error should not be NotFound type")
	}
}

func TestBuildInternalDomain_LengthValidation(t *testing.T) {
	cases := []struct {
		name          string
		subdomain     string
		clusterDomain string
		wantReason    string
	}{
		{
			name:          "at limit passes",
			subdomain:     "a",
			clusterDomain: strings.Repeat("b", 62), // 1 + 1 + 62 = 64
			wantReason:    "",
		},
		{
			name:          "over limit fails",
			subdomain:     "a",
			clusterDomain: strings.Repeat("b", 63), // 1 + 1 + 63 = 65
			wantReason:    externalv1alpha1.ReasonInternalDomainTooLong,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInternalDomainLength(tc.subdomain + "." + tc.clusterDomain)
			if tc.wantReason == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error with reason %s, got nil", tc.wantReason)
			}
			reason, ok := externalgateway.ErrorReason(err)
			if !ok || reason != tc.wantReason {
				t.Fatalf("expected reason %s, got reason=%q ok=%v", tc.wantReason, reason, ok)
			}
		})
	}
}
