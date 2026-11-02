package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConditionConstructors(t *testing.T) {
	const msg = "test message"

	tests := []struct {
		name       string
		got        metav1.Condition
		wantType   string
		wantStatus metav1.ConditionStatus
		wantReason string
	}{
		// Ready
		{name: "ProcessingCondition", got: ProcessingCondition(), wantType: ConditionTypeReady, wantStatus: metav1.ConditionUnknown, wantReason: ReasonReconcileProcessing},
		{name: "ReadyCondition", got: ReadyCondition(), wantType: ConditionTypeReady, wantStatus: metav1.ConditionTrue, wantReason: ReasonReconcileSucceeded},
		{name: "ErrorCondition", got: ErrorCondition(ReasonReconcileFailed, msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonReconcileFailed},
		{name: "DeletionBlockedExistingResourcesCondition", got: DeletionBlockedExistingResourcesCondition(msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonDeletionBlockedExistingResources},
		// Oathkeeper (reuses ConditionTypeReady)
		{name: "OathkeeperReadyCondition", got: OathkeeperReadyCondition(), wantType: ConditionTypeReady, wantStatus: metav1.ConditionTrue, wantReason: ReasonOathkeeperReconcileSucceeded},
		{name: "OathkeeperDisabledCondition", got: OathkeeperDisabledCondition(), wantType: ConditionTypeReady, wantStatus: metav1.ConditionTrue, wantReason: ReasonOathkeeperReconcileDisabled},
		{name: "OathkeeperErrorCondition", got: OathkeeperErrorCondition(msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonOathkeeperReconcileFailed},
		// KymaGateway
		{name: "KymaGatewayReadyCondition", got: KymaGatewayReadyCondition(), wantType: ConditionTypeKymaGateway, wantStatus: metav1.ConditionTrue, wantReason: ReasonKymaGatewayReconcileSucceeded},
		{name: "KymaGatewayDisabledCondition", got: KymaGatewayDisabledCondition(), wantType: ConditionTypeKymaGateway, wantStatus: metav1.ConditionFalse, wantReason: ReasonKymaGatewayReconcileDisabled},
		{name: "KymaGatewayProcessingCondition", got: KymaGatewayProcessingCondition(), wantType: ConditionTypeKymaGateway, wantStatus: metav1.ConditionUnknown, wantReason: ReasonKymaGatewayReconcileProcessing},
		{name: "KymaGatewayErrorCondition", got: KymaGatewayErrorCondition(msg), wantType: ConditionTypeKymaGateway, wantStatus: metav1.ConditionFalse, wantReason: ReasonKymaGatewayReconcileFailed},
		{name: "KymaGatewayDeletionBlockedCondition", got: KymaGatewayDeletionBlockedCondition(msg), wantType: ConditionTypeKymaGateway, wantStatus: metav1.ConditionFalse, wantReason: ReasonKymaGatewayDeletionBlocked},
		// DNSEntry
		{name: "DNSEntryReadyCondition", got: DNSEntryReadyCondition(), wantType: ConditionTypeDNSEntry, wantStatus: metav1.ConditionTrue, wantReason: ReasonDNSEntryReconcileSucceeded},
		{name: "DNSEntryProcessingCondition", got: DNSEntryProcessingCondition(), wantType: ConditionTypeDNSEntry, wantStatus: metav1.ConditionUnknown, wantReason: ReasonDNSEntryReconcilePending},
		{name: "DNSEntryErrorCondition", got: DNSEntryErrorCondition(msg), wantType: ConditionTypeDNSEntry, wantStatus: metav1.ConditionFalse, wantReason: ReasonDNSEntryReconcileFailed},
		// Certificate
		{name: "CertificateReadyCondition", got: CertificateReadyCondition(), wantType: ConditionTypeCertificate, wantStatus: metav1.ConditionTrue, wantReason: ReasonCertificateReconcileSucceeded},
		{name: "CertificateProcessingCondition", got: CertificateProcessingCondition(), wantType: ConditionTypeCertificate, wantStatus: metav1.ConditionUnknown, wantReason: ReasonCertificateReconcilePending},
		{name: "CertificateErrorCondition", got: CertificateErrorCondition(msg), wantType: ConditionTypeCertificate, wantStatus: metav1.ConditionFalse, wantReason: ReasonCertificateReconcileFailed},
		// NetworkPolicy
		{name: "NetworkPolicyReadyCondition", got: NetworkPolicyReadyCondition(), wantType: ConditionTypeNetworkPolicy, wantStatus: metav1.ConditionTrue, wantReason: ReasonNetworkPolicyReconcileSucceeded},
		{name: "NetworkPolicyDisabledCondition", got: NetworkPolicyDisabledCondition(), wantType: ConditionTypeNetworkPolicy, wantStatus: metav1.ConditionFalse, wantReason: ReasonNetworkPolicyReconcileDisabled},
		{name: "NetworkPolicyErrorCondition", got: NetworkPolicyErrorCondition(msg), wantType: ConditionTypeNetworkPolicy, wantStatus: metav1.ConditionFalse, wantReason: ReasonNetworkPolicyReconcileFailed},
		// Dependencies
		{name: "DependenciesReadyCondition", got: DependenciesReadyCondition(), wantType: ConditionTypeDependencies, wantStatus: metav1.ConditionTrue, wantReason: ReasonDependenciesReconcileSucceeded},
		{name: "DependenciesMissingCondition", got: DependenciesMissingCondition(msg), wantType: ConditionTypeDependencies, wantStatus: metav1.ConditionFalse, wantReason: ReasonDependenciesMissing},
		{name: "DependenciesErrorCondition", got: DependenciesErrorCondition(msg), wantType: ConditionTypeDependencies, wantStatus: metav1.ConditionFalse, wantReason: ReasonDependenciesReconcileFailed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.Type != tc.wantType {
				t.Errorf("Type: got %q, want %q", tc.got.Type, tc.wantType)
			}
			if tc.got.Status != tc.wantStatus {
				t.Errorf("Status: got %q, want %q", tc.got.Status, tc.wantStatus)
			}
			if tc.got.Reason != tc.wantReason {
				t.Errorf("Reason: got %q, want %q", tc.got.Reason, tc.wantReason)
			}
			if tc.got.ObservedGeneration != 0 {
				t.Errorf("ObservedGeneration: got %d, want 0 (stamped later by UpdateApiGatewayStatus)", tc.got.ObservedGeneration)
			}
		})
	}
}
