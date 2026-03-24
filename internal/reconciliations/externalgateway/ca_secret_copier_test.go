package externalgateway

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

// newSecretRef creates a SecretReference for testing
func newSecretRef(name, namespace string) *corev1.SecretReference {
	return &corev1.SecretReference{
		Name:      name,
		Namespace: namespace,
	}
}

func TestReconcileCASecret(t *testing.T) {
	tests := []struct {
		name               string
		externalSpec       externalv1alpha1.ExternalGatewaySpec
		sourceSecretData   map[string][]byte
		sourceSecretExists bool
		targetSecretExists bool
		targetSecretData   map[string][]byte
		expectError        bool
		errorContains      string
		expectCreate       bool
		expectUpdate       bool
	}{
		{
			name: "source secret exists with ca.crt key - creates target",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("source-ca-secret", ""),
			},
			sourceSecretData: map[string][]byte{
				"ca.crt": []byte("-----BEGIN CERTIFICATE-----\ntest-ca-cert\n-----END CERTIFICATE-----"),
			},
			sourceSecretExists: true,
			targetSecretExists: false,
			expectError:        false,
			expectCreate:       true,
		},
		{
			name: "source secret with single key (different name) - uses it automatically",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("source-ca-secret", ""),
			},
			sourceSecretData: map[string][]byte{
				"root-ca.crt": []byte("-----BEGIN CERTIFICATE-----\ntest-ca-cert\n-----END CERTIFICATE-----"),
			},
			sourceSecretExists: true,
			targetSecretExists: false,
			expectError:        false,
			expectCreate:       true,
		},
		{
			name: "source secret missing ca.crt key - returns error",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("source-ca-secret", ""),
			},
			sourceSecretData: map[string][]byte{
				"other-key":  []byte("some-value"),
				"other-key2": []byte("some-value2"),
			},
			sourceSecretExists: true,
			targetSecretExists: false,
			expectError:        true,
			errorContains:      "does not contain 'ca.crt' key",
		},
		{
			name: "source secret is empty - returns error",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("source-ca-secret", ""),
			},
			sourceSecretData:   map[string][]byte{},
			sourceSecretExists: true,
			targetSecretExists: false,
			expectError:        true,
			errorContains:      "is empty",
		},
		{
			name: "source secret with multiple keys including ca.crt - uses ca.crt",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("source-ca-secret", ""),
			},
			sourceSecretData: map[string][]byte{
				"ca.crt":     []byte("-----BEGIN CERTIFICATE-----\ncorrect-ca-cert\n-----END CERTIFICATE-----"),
				"other-key":  []byte("some-other-data"),
				"other-key2": []byte("more-data"),
			},
			sourceSecretExists: true,
			targetSecretExists: false,
			expectError:        false,
			expectCreate:       true,
		},
		{
			name: "source secret not found - returns error",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("missing-secret", ""),
			},
			sourceSecretExists: false,
			expectError:        true,
			errorContains:      "failed to get source CA secret",
		},
		{
			name: "target secret already exists - updates data and labels",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("source-ca-secret", ""),
			},
			sourceSecretData: map[string][]byte{
				"ca.crt": []byte("-----BEGIN CERTIFICATE-----\nnew-ca-cert\n-----END CERTIFICATE-----"),
			},
			sourceSecretExists: true,
			targetSecretExists: true,
			targetSecretData: map[string][]byte{
				"ca.crt": []byte("-----BEGIN CERTIFICATE-----\nold-ca-cert\n-----END CERTIFICATE-----"),
			},
			expectError:  false,
			expectUpdate: true,
		},
		{
			name: "correct naming convention - gateway-name-cacert",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{
				CASecretRef: newSecretRef("source-ca-secret", ""),
			},
			sourceSecretData: map[string][]byte{
				"ca.crt": []byte("-----BEGIN CERTIFICATE-----\ntest-ca-cert\n-----END CERTIFICATE-----"),
			},
			sourceSecretExists: true,
			targetSecretExists: false,
			expectError:        false,
			expectCreate:       true,
		},
		{
			name: "cross-namespace CA secret reference",
			externalSpec: externalv1alpha1.ExternalGatewaySpec{

				CASecretRef: newSecretRef("source-ca-secret", "custom-namespace"),
			},
			sourceSecretData: map[string][]byte{
				"ca.crt": []byte("-----BEGIN CERTIFICATE-----\ntest-ca-cert\n-----END CERTIFICATE-----"),
			},
			sourceSecretExists: true,
			targetSecretExists: false,
			expectError:        false,
			expectCreate:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create scheme
			scheme := runtime.NewScheme()
			_ = corev1.AddToScheme(scheme)
			_ = externalv1alpha1.AddToScheme(scheme)

			// Create fake client
			clientBuilder := fake.NewClientBuilder().WithScheme(scheme)

			// Add source secret if it should exist
			if tt.sourceSecretExists {
				// Determine namespace: use explicit namespace or default to test-namespace
				sourceNamespace := tt.externalSpec.CASecretRef.Namespace
				if sourceNamespace == "" {
					sourceNamespace = "test-namespace"
				}

				sourceSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tt.externalSpec.CASecretRef.Name,
						Namespace: sourceNamespace,
					},
					Data: tt.sourceSecretData,
				}
				clientBuilder = clientBuilder.WithObjects(sourceSecret)
			}

			// Add target secret if it should exist
			if tt.targetSecretExists {
				targetSecretName := "test-gateway-gateway-tls-cacert"
				targetSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      targetSecretName,
						Namespace: istioSystemNamespace,
					},
					Data: tt.targetSecretData,
				}
				clientBuilder = clientBuilder.WithObjects(targetSecret)
			}

			fakeClient := clientBuilder.Build()

			// Create ExternalGateway
			external := &externalv1alpha1.ExternalGateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gateway",
					Namespace: "test-namespace",
				},
				Spec: tt.externalSpec,
			}

			// Execute
			ctx := context.Background()
			err := ReconcileCASecret(ctx, fakeClient, external)

			// Assert error
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error message '%s' does not contain '%s'", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify target secret was created/updated
			targetSecretName := "test-gateway-gateway-tls-cacert"
			targetSecret := &corev1.Secret{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Name:      targetSecretName,
				Namespace: istioSystemNamespace,
			}, targetSecret)

			if err != nil {
				t.Errorf("failed to get target secret: %v", err)
				return
			}

			// Verify ca.crt data
			if _, exists := targetSecret.Data["ca.crt"]; !exists {
				t.Errorf("target secret missing 'ca.crt' key")
			}

			// Verify data matches source (source could have any key if single key)
			var sourceData []byte
			if len(tt.sourceSecretData) == 1 {
				for _, data := range tt.sourceSecretData {
					sourceData = data
					break
				}
			} else {
				sourceData = tt.sourceSecretData["ca.crt"]
			}

			if string(targetSecret.Data["ca.crt"]) != string(sourceData) {
				t.Errorf("target ca.crt data does not match source")
			}

			// Verify labels
			expectedLabels := map[string]string{
				"app.kubernetes.io/managed-by":                      "externalgateway-controller",
				"externalgateway.gateway.kyma-project.io/name":      "test-gateway",
				"externalgateway.gateway.kyma-project.io/namespace": "test-namespace",
			}
			for key, expectedValue := range expectedLabels {
				if actualValue, exists := targetSecret.Labels[key]; !exists {
					t.Errorf("missing label '%s'", key)
				} else if actualValue != expectedValue {
					t.Errorf("label '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
				}
			}
		})
	}
}
