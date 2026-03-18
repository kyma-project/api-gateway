package external_test

import (
	"context"
	"testing"
	"time"

	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

func TestExternalGatewayCreation(t *testing.T) {

	// Create regions ConfigMap
	regionsConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "external-gateway-regions",
			Namespace: testNamespace,
		},
		Data: map[string]string{
			"regions.yaml": `
- Provider: aws
  Region: us-east-1
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, OU=test-uuid-1, L=gateway, CN=aws/us-east-1"
- Provider: gcp
  Region: europe-west1
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, OU=test-uuid-2, L=gateway, CN=gcp/europe-west1"
`,
		},
	}
	if err := k8sClient.Create(ctx, regionsConfigMap); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create regions ConfigMap: %v", err)
	}

	// Create CA Secret in application namespace
	caSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ca-secret",
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"cacert": []byte("-----BEGIN CERTIFICATE-----\ntest-ca-certificate\n-----END CERTIFICATE-----"),
		},
	}
	if err := k8sClient.Create(ctx, caSecret); err != nil {
		t.Fatalf("failed to create CA secret: %v", err)
	}
	defer k8sClient.Delete(ctx, caSecret)

	// Create ExternalGateway
	externalGateway := &externalv1alpha1.ExternalGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-external-gateway",
			Namespace: testNamespace,
		},
		Spec: externalv1alpha1.ExternalGatewaySpec{
			ExternalDomain: "api.customer.com",
			InternalDomain: externalv1alpha1.InternalDomainConfig{
				KymaSubdomain: "test-gateway",
			},
			Regions: []string{
				"aws/us-east-1",
				"gcp/europe-west1",
			},
			Gateway: "test-gateway",
			CASecretRef: &corev1.SecretReference{
				Name:      "test-ca-secret",
				Namespace: "",
			},
		},
	}
	if err := k8sClient.Create(ctx, externalGateway); err != nil {
		t.Fatalf("failed to create ExternalGateway: %v", err)
	}
	defer k8sClient.Delete(ctx, externalGateway)

	// Wait for ExternalGateway to be created
	externalGatewayLookupKey := types.NamespacedName{
		Name:      "test-external-gateway",
		Namespace: testNamespace,
	}
	createdExternalGateway := &externalv1alpha1.ExternalGateway{}

	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, externalGatewayLookupKey, createdExternalGateway)
		return err == nil
	}); err != nil {
		t.Fatalf("ExternalGateway was not created: %v", err)
	}

	// Verify finalizer was added
	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, externalGatewayLookupKey, createdExternalGateway)
		return err == nil && len(createdExternalGateway.Finalizers) > 0
	}); err != nil {
		t.Fatalf("Finalizer was not added: %v", err)
	}

	if !containsString(createdExternalGateway.Finalizers, "externalgateways.gateway.kyma-project.io/finalizer") {
		t.Errorf("Expected finalizer not found, got: %v", createdExternalGateway.Finalizers)
	}

	// Verify CA Secret was copied to istio-system
	caSecretCopyLookupKey := types.NamespacedName{
		Name:      "test-gateway-cacert",
		Namespace: istioSystemNs,
	}
	caSecretCopy := &corev1.Secret{}

	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, caSecretCopyLookupKey, caSecretCopy)
		return err == nil
	}); err != nil {
		t.Fatalf("CA Secret was not copied to istio-system: %v", err)
	}
	defer k8sClient.Delete(ctx, caSecretCopy)

	if _, exists := caSecretCopy.Data["cacert"]; !exists {
		t.Error("CA Secret copy does not contain 'cacert' key")
	}

	if string(caSecretCopy.Data["cacert"]) != string(caSecret.Data["cacert"]) {
		t.Error("CA Secret copy data does not match source")
	}

	if caSecretCopy.Labels["app.kubernetes.io/managed-by"] != "externalgateway-controller" {
		t.Errorf("CA Secret copy has wrong managed-by label: %v", caSecretCopy.Labels)
	}

	// Verify Istio Gateway was created
	istioGatewayLookupKey := types.NamespacedName{
		Name:      "test-gateway",
		Namespace: testNamespace,
	}
	istioGateway := &networkingv1beta1.Gateway{}

	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, istioGatewayLookupKey, istioGateway)
		return err == nil
	}); err != nil {
		t.Fatalf("Istio Gateway was not created: %v", err)
	}
	defer k8sClient.Delete(ctx, istioGateway)

	if len(istioGateway.Spec.Servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(istioGateway.Spec.Servers))
	}

	if istioGateway.Spec.Servers[0].Port.Protocol != "HTTPS" {
		t.Errorf("Expected HTTPS protocol, got %s", istioGateway.Spec.Servers[0].Port.Protocol)
	}

	if istioGateway.Spec.Servers[0].Port.Number != 443 {
		t.Errorf("Expected port 443, got %d", istioGateway.Spec.Servers[0].Port.Number)
	}

	if istioGateway.Spec.Servers[0].Tls.Mode.String() != "MUTUAL" {
		t.Errorf("Expected MUTUAL TLS mode, got %s", istioGateway.Spec.Servers[0].Tls.Mode.String())
	}

	if !containsString(istioGateway.Spec.Servers[0].Hosts, "api.customer.com") {
		t.Errorf("Expected host 'api.customer.com' not found in: %v", istioGateway.Spec.Servers[0].Hosts)
	}

	if istioGateway.Labels["app.kubernetes.io/managed-by"] != "externalgateway-controller" {
		t.Errorf("Gateway has wrong managed-by label: %v", istioGateway.Labels)
	}

	// Verify EnvoyFilters were created
	envoyFilterList := &networkingv1alpha3.EnvoyFilterList{}
	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.List(ctx, envoyFilterList, client.InNamespace(istioSystemNs))
		if err != nil {
			return false
		}
		count := 0
		for _, ef := range envoyFilterList.Items {
			if ef.Labels["externalgateway.gateway.kyma-project.io/name"] == "test-external-gateway" {
				count++
			}
		}
		return count == 2 // xfcc-sanitization and cert-validation
	}); err != nil {
		t.Fatalf("EnvoyFilters were not created: %v", err)
	}

	// Cleanup EnvoyFilters
	for i := range envoyFilterList.Items {
		if envoyFilterList.Items[i].Labels["externalgateway.gateway.kyma-project.io/name"] == "test-external-gateway" {
			k8sClient.Delete(ctx, envoyFilterList.Items[i])
		}
	}

	// Verify status was updated to Ready
	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, externalGatewayLookupKey, createdExternalGateway)
		return err == nil && createdExternalGateway.Status.State == externalv1alpha1.Ready
	}); err != nil {
		t.Fatalf("Status was not updated to Ready: %v, state: %s", err, createdExternalGateway.Status.State)
	}

	// Test deletion
	if err := k8sClient.Delete(ctx, externalGateway); err != nil {
		t.Fatalf("failed to delete ExternalGateway: %v", err)
	}

	// Verify resources are cleaned up
	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, caSecretCopyLookupKey, caSecretCopy)
		return apierrors.IsNotFound(err)
	}); err != nil {
		t.Errorf("CA Secret copy was not deleted: %v", err)
	}

	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, istioGatewayLookupKey, istioGateway)
		return apierrors.IsNotFound(err)
	}); err != nil {
		t.Errorf("Istio Gateway was not deleted: %v", err)
	}
}

func TestExternalGatewayMissingCASecret(t *testing.T) {

	// Create regions ConfigMap
	regionsConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "external-gateway-regions",
			Namespace: testNamespace,
		},
		Data: map[string]string{
			"regions.yaml": `
- Provider: aws
  Region: us-east-1
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, OU=test-uuid, L=gateway, CN=aws/us-east-1"
`,
		},
	}
	if err := k8sClient.Create(ctx, regionsConfigMap); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create regions ConfigMap: %v", err)
	}

	// Create ExternalGateway without CA Secret
	externalGateway := &externalv1alpha1.ExternalGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-missing-ca-secret",
			Namespace: testNamespace,
		},
		Spec: externalv1alpha1.ExternalGatewaySpec{
			ExternalDomain: "api2.customer.com",
			InternalDomain: externalv1alpha1.InternalDomainConfig{
				KymaSubdomain: "test-gateway-2",
			},
			Regions: []string{
				"aws/us-east-1",
			},
			Gateway: "test-gateway-2",
			CASecretRef: &corev1.SecretReference{
				Name:      "missing-ca-secret",
				Namespace: "",
			},
		},
	}
	if err := k8sClient.Create(ctx, externalGateway); err != nil {
		t.Fatalf("failed to create ExternalGateway: %v", err)
	}
	defer k8sClient.Delete(ctx, externalGateway)

	// Wait for ExternalGateway to be created
	externalGatewayLookupKey := types.NamespacedName{
		Name:      "test-missing-ca-secret",
		Namespace: testNamespace,
	}
	createdExternalGateway := &externalv1alpha1.ExternalGateway{}

	// Verify status is set to Error
	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, externalGatewayLookupKey, createdExternalGateway)
		return err == nil && createdExternalGateway.Status.State == externalv1alpha1.Error
	}); err != nil {
		t.Fatalf("Status was not updated to Error: %v, state: %s", err, createdExternalGateway.Status.State)
	}

	if !stringContains(createdExternalGateway.Status.Description, "failed to get source CA secret") {
		t.Errorf("Expected error message not found, got: %s", createdExternalGateway.Status.Description)
	}
}

func TestExternalGatewayInvalidCASecret(t *testing.T) {

	// Create regions ConfigMap
	regionsConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "external-gateway-regions",
			Namespace: testNamespace,
		},
		Data: map[string]string{
			"regions.yaml": `
- Provider: aws
  Region: us-east-1
  CertSubjects:
    - "C=US, O=Example Inc, OU=Clients, OU=test-uuid, L=gateway, CN=aws/us-east-1"
`,
		},
	}
	if err := k8sClient.Create(ctx, regionsConfigMap); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create regions ConfigMap: %v", err)
	}

	// Create CA Secret without cacert key
	invalidSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "invalid-ca-secret",
			Namespace: testNamespace,
		},
		Data: map[string][]byte{
			"other-key": []byte("some-value"),
		},
	}
	if err := k8sClient.Create(ctx, invalidSecret); err != nil {
		t.Fatalf("failed to create invalid secret: %v", err)
	}
	defer k8sClient.Delete(ctx, invalidSecret)

	// Create ExternalGateway
	externalGateway := &externalv1alpha1.ExternalGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-invalid-ca-secret",
			Namespace: testNamespace,
		},
		Spec: externalv1alpha1.ExternalGatewaySpec{
			ExternalDomain: "api3.customer.com",
			InternalDomain: externalv1alpha1.InternalDomainConfig{
				KymaSubdomain: "test-gateway-3",
			},
			Regions: []string{
				"aws/us-east-1",
			},
			Gateway: "test-gateway-3",
			CASecretRef: &corev1.SecretReference{
				Name:      "invalid-ca-secret",
				Namespace: "",
			},
		},
	}
	if err := k8sClient.Create(ctx, externalGateway); err != nil {
		t.Fatalf("failed to create ExternalGateway: %v", err)
	}
	defer k8sClient.Delete(ctx, externalGateway)

	// Wait for ExternalGateway to be created
	externalGatewayLookupKey := types.NamespacedName{
		Name:      "test-invalid-ca-secret",
		Namespace: testNamespace,
	}
	createdExternalGateway := &externalv1alpha1.ExternalGateway{}

	// Verify status is set to Error
	if err := waitForCondition(t, 10*time.Second, func() bool {
		err := k8sClient.Get(ctx, externalGatewayLookupKey, createdExternalGateway)
		return err == nil && createdExternalGateway.Status.State == externalv1alpha1.Error
	}); err != nil {
		t.Fatalf("Status was not updated to Error: %v, state: %s", err, createdExternalGateway.Status.State)
	}

	if !stringContains(createdExternalGateway.Status.Description, "does not contain 'cacert' key") {
		t.Errorf("Expected error message not found, got: %s", createdExternalGateway.Status.Description)
	}
}

// Helper functions

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) error {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return context.DeadlineExceeded
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
