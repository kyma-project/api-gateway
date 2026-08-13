package external_test

import (
	"testing"
	"time"

	certv1alpha1 "github.com/gardener/cert-management/pkg/apis/cert/v1alpha1"
	dnsv1alpha1 "github.com/gardener/external-dns-management/pkg/apis/dns/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

// TestGardenerAvailable_DNSTerminalError verifies that when the DNSEntry enters a terminal
// error state, both the DNSEntryReady AND CertificateReady conditions are still written
// (i.e. the controller does not short-circuit before reconciling Certificate).
func TestGardenerAvailable_DNSTerminalError(t *testing.T) {
	extGatewayName := "test-dns-terminal-error"

	regionsConfigMap := getRegionsConfigMap()
	if err := k8sClient.Create(ctx, regionsConfigMap); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create regions ConfigMap: %v", err)
	}

	caSecret := getCASecret("ca-secret-dns-error")
	if err := k8sClient.Create(ctx, caSecret); err != nil {
		t.Fatalf("failed to create CA secret: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, caSecret) }()

	externalGateway := getExternalGateway(extGatewayName, "ca-secret-dns-error")
	if err := k8sClient.Create(ctx, externalGateway); err != nil {
		t.Fatalf("failed to create ExternalGateway: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, externalGateway) }()

	lookupKey := types.NamespacedName{Name: extGatewayName, Namespace: testNamespace}
	eg := &externalv1alpha1.ExternalGateway{}

	// Wait for the controller to create the DNSEntry.
	dnsKey := types.NamespacedName{Name: externalGateway.DNSEntryName(), Namespace: istioSystemNs}
	dnsEntry := &dnsv1alpha1.DNSEntry{}
	if err := waitForCondition(t, 15*time.Second, func() bool {
		return k8sClient.Get(ctx, dnsKey, dnsEntry) == nil
	}); err != nil {
		t.Fatalf("DNSEntry was not created: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, dnsEntry) }()

	// Patch DNSEntry status to a terminal error state.
	dnsEntry.Status.State = dnsv1alpha1.STATE_ERROR
	dnsEntry.Status.Message = ptr.To("simulated DNS error")
	if err := k8sClient.Status().Update(ctx, dnsEntry); err != nil {
		t.Fatalf("failed to patch DNSEntry status: %v", err)
	}

	// Wait for ExternalGateway to reach Error state.
	if err := waitForCondition(t, 15*time.Second, func() bool {
		if err := k8sClient.Get(ctx, lookupKey, eg); err != nil {
			return false
		}
		return eg.Status.State == externalv1alpha1.Error
	}); err != nil {
		t.Fatalf("ExternalGateway did not reach Error state: %v, state: %s", err, eg.Status.State)
	}

	// Both conditions must be present — this is the core assertion for the fix.
	assertCondition(t, eg, externalv1alpha1.ConditionTypeDNSEntryReady, metav1.ConditionFalse, externalv1alpha1.ReasonDNSEntryError)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeCertificateReady, metav1.ConditionFalse, externalv1alpha1.ReasonCertificatePending)

	// Clean up Certificate created by the controller.
	cert := &certv1alpha1.Certificate{}
	certKey := types.NamespacedName{Name: externalGateway.CertificateName(), Namespace: istioSystemNs}
	if k8sClient.Get(ctx, certKey, cert) == nil {
		_ = k8sClient.Delete(ctx, cert)
	}
}

// TestGardenerAvailable_CertTerminalError verifies that when the Certificate enters a
// terminal error state, both conditions are still present in the status.
func TestGardenerAvailable_CertTerminalError(t *testing.T) {
	extGatewayName := "test-cert-terminal-error"

	regionsConfigMap := getRegionsConfigMap()
	if err := k8sClient.Create(ctx, regionsConfigMap); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create regions ConfigMap: %v", err)
	}

	caSecret := getCASecret("ca-secret-cert-error")
	if err := k8sClient.Create(ctx, caSecret); err != nil {
		t.Fatalf("failed to create CA secret: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, caSecret) }()

	externalGateway := getExternalGateway(extGatewayName, "ca-secret-cert-error")
	if err := k8sClient.Create(ctx, externalGateway); err != nil {
		t.Fatalf("failed to create ExternalGateway: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, externalGateway) }()

	lookupKey := types.NamespacedName{Name: extGatewayName, Namespace: testNamespace}
	eg := &externalv1alpha1.ExternalGateway{}

	// Wait for the controller to create both DNSEntry and Certificate.
	dnsKey := types.NamespacedName{Name: externalGateway.DNSEntryName(), Namespace: istioSystemNs}
	dnsEntry := &dnsv1alpha1.DNSEntry{}
	if err := waitForCondition(t, 15*time.Second, func() bool {
		return k8sClient.Get(ctx, dnsKey, dnsEntry) == nil
	}); err != nil {
		t.Fatalf("DNSEntry was not created: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, dnsEntry) }()

	certKey := types.NamespacedName{Name: externalGateway.CertificateName(), Namespace: istioSystemNs}
	cert := &certv1alpha1.Certificate{}
	if err := waitForCondition(t, 15*time.Second, func() bool {
		return k8sClient.Get(ctx, certKey, cert) == nil
	}); err != nil {
		t.Fatalf("Certificate was not created: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, cert) }()

	// Drive DNS to Ready so only the Certificate path is the error source.
	dnsEntry.Status.State = dnsv1alpha1.STATE_READY
	if err := k8sClient.Status().Update(ctx, dnsEntry); err != nil {
		t.Fatalf("failed to patch DNSEntry status to Ready: %v", err)
	}

	// Patch Certificate status to a terminal error state.
	msg := "simulated cert error"
	cert.Status.State = certv1alpha1.StateError
	cert.Status.Message = &msg
	if err := k8sClient.Status().Update(ctx, cert); err != nil {
		t.Fatalf("failed to patch Certificate status: %v", err)
	}

	// Wait for ExternalGateway to reach Error state.
	if err := waitForCondition(t, 15*time.Second, func() bool {
		if err := k8sClient.Get(ctx, lookupKey, eg); err != nil {
			return false
		}
		return eg.Status.State == externalv1alpha1.Error
	}); err != nil {
		t.Fatalf("ExternalGateway did not reach Error state: %v, state: %s", err, eg.Status.State)
	}

	assertCondition(t, eg, externalv1alpha1.ConditionTypeDNSEntryReady, metav1.ConditionTrue, externalv1alpha1.ReasonReady)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeCertificateReady, metav1.ConditionFalse, externalv1alpha1.ReasonCertificateError)
}

// TestGardenerAvailable_BothReady verifies the happy path: when both DNSEntry and Certificate
// reach Ready state the ExternalGateway transitions to Ready with all conditions True.
func TestGardenerAvailable_BothReady(t *testing.T) {
	extGatewayName := "test-gardener-both-ready"

	regionsConfigMap := getRegionsConfigMap()
	if err := k8sClient.Create(ctx, regionsConfigMap); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create regions ConfigMap: %v", err)
	}

	caSecret := getCASecret("ca-secret-both-ready")
	if err := k8sClient.Create(ctx, caSecret); err != nil {
		t.Fatalf("failed to create CA secret: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, caSecret) }()

	externalGateway := getExternalGateway(extGatewayName, "ca-secret-both-ready")
	if err := k8sClient.Create(ctx, externalGateway); err != nil {
		t.Fatalf("failed to create ExternalGateway: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, externalGateway) }()

	lookupKey := types.NamespacedName{Name: extGatewayName, Namespace: testNamespace}
	eg := &externalv1alpha1.ExternalGateway{}

	// Wait for both sub-resources to be created by the controller.
	dnsKey := types.NamespacedName{Name: externalGateway.DNSEntryName(), Namespace: istioSystemNs}
	dnsEntry := &dnsv1alpha1.DNSEntry{}
	if err := waitForCondition(t, 15*time.Second, func() bool {
		return k8sClient.Get(ctx, dnsKey, dnsEntry) == nil
	}); err != nil {
		t.Fatalf("DNSEntry was not created: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, dnsEntry) }()

	certKey := types.NamespacedName{Name: externalGateway.CertificateName(), Namespace: istioSystemNs}
	cert := &certv1alpha1.Certificate{}
	if err := waitForCondition(t, 15*time.Second, func() bool {
		return k8sClient.Get(ctx, certKey, cert) == nil
	}); err != nil {
		t.Fatalf("Certificate was not created: %v", err)
	}
	defer func() {
		_ = k8sClient.Delete(ctx, cert)
		tlsSecret := externalGateway.TLSSecretName()
		_ = k8sClient.Delete(ctx, &certv1alpha1.Certificate{ObjectMeta: metav1.ObjectMeta{Name: tlsSecret, Namespace: istioSystemNs}})
	}()

	dnsEntry.Status.State = dnsv1alpha1.STATE_READY
	if err := k8sClient.Status().Update(ctx, dnsEntry); err != nil {
		t.Fatalf("failed to patch DNSEntry status to Ready: %v", err)
	}

	msg := ""
	cert.Status.State = certv1alpha1.StateReady
	cert.Status.Message = &msg
	if err := k8sClient.Status().Update(ctx, cert); err != nil {
		t.Fatalf("failed to patch Certificate status to Ready: %v", err)
	}

	if err := waitForCondition(t, 15*time.Second, func() bool {
		if err := k8sClient.Get(ctx, lookupKey, eg); err != nil {
			return false
		}
		return eg.Status.State == externalv1alpha1.Ready
	}); err != nil {
		t.Fatalf("ExternalGateway did not reach Ready state: %v, state: %s", err, eg.Status.State)
	}

	assertCondition(t, eg, externalv1alpha1.ConditionTypeDNSEntryReady, metav1.ConditionTrue, externalv1alpha1.ReasonReady)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeCertificateReady, metav1.ConditionTrue, externalv1alpha1.ReasonReady)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeGatewayConfigured, metav1.ConditionTrue, externalv1alpha1.ReasonReady)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeReady, metav1.ConditionTrue, externalv1alpha1.ReasonReady)

	// Clean up EnvoyFilters
	efList := listEnvoyFiltersForGateway(t, externalGateway)
	for i := range efList {
		_ = k8sClient.Delete(ctx, efList[i], &client.DeleteOptions{})
	}
}
