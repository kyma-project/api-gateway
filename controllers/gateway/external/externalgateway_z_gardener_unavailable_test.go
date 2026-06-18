package external_test

// TestGardenerUnavailable is intentionally in a file prefixed with "z_" so that Go runs it
// last within this package. It deletes the Gardener CRDs from the envtest cluster, which is
// irreversible within a single test run. All other tests must complete before this one fires.

import (
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	externalv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/external/v1alpha1"
)

// TestGardenerUnavailable verifies that when Gardener CRDs are not present, both
// DNSEntryReady and CertificateReady conditions are set to False with GardenerCRDUnavailable
// reason, and the ExternalGateway still reaches Ready (Istio resources are reconciled).
func TestGardenerUnavailable(t *testing.T) {
	deleteGardenerCRDs(t)

	extGatewayName := "test-gardener-unavailable"

	regionsConfigMap := getRegionsConfigMap()
	if err := k8sClient.Create(ctx, regionsConfigMap); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create regions ConfigMap: %v", err)
	}

	caSecret := getCASecret("ca-secret-unavailable")
	if err := k8sClient.Create(ctx, caSecret); err != nil {
		t.Fatalf("failed to create CA secret: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, caSecret) }()

	externalGateway := getExternalGateway(extGatewayName, "ca-secret-unavailable")
	if err := k8sClient.Create(ctx, externalGateway); err != nil {
		t.Fatalf("failed to create ExternalGateway: %v", err)
	}
	defer func() { _ = k8sClient.Delete(ctx, externalGateway) }()

	lookupKey := types.NamespacedName{Name: extGatewayName, Namespace: testNamespace}
	eg := &externalv1alpha1.ExternalGateway{}

	if err := waitForCondition(t, 15*time.Second, func() bool {
		if err := k8sClient.Get(ctx, lookupKey, eg); err != nil {
			return false
		}
		return eg.Status.State == externalv1alpha1.Ready
	}); err != nil {
		t.Fatalf("ExternalGateway did not reach Ready state: %v, state: %s", err, eg.Status.State)
	}

	assertCondition(t, eg, externalv1alpha1.ConditionTypeDNSEntryReady, metav1.ConditionFalse, externalv1alpha1.ReasonGardenerCRDUnavailable)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeCertificateReady, metav1.ConditionFalse, externalv1alpha1.ReasonGardenerCRDUnavailable)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeGatewayConfigured, metav1.ConditionTrue, externalv1alpha1.ReasonReady)
	assertCondition(t, eg, externalv1alpha1.ConditionTypeReady, metav1.ConditionTrue, externalv1alpha1.ReasonReady)

	// Clean up EnvoyFilters
	efList := listEnvoyFiltersForGateway(t, externalGateway)
	for i := range efList {
		_ = k8sClient.Delete(ctx, efList[i], &client.DeleteOptions{})
	}
}
