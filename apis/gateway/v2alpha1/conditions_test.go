package v2alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConditionConstructors(t *testing.T) {
	const gen = int64(17)
	const msg = "some message"

	tests := []struct {
		name       string
		got        metav1.Condition
		wantType   string
		wantStatus metav1.ConditionStatus
		wantReason string
		wantMsg    string // empty means "don't assert message"
	}{
		// Ready
		{name: "ReadyCondition", got: ReadyCondition(gen), wantType: ConditionTypeReady, wantStatus: metav1.ConditionTrue, wantReason: ReasonReconcileSucceeded},
		{name: "ErrorCondition", got: ErrorCondition(gen, ReasonReconcileFailed, msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonReconcileFailed, wantMsg: msg},
		{name: "GatewayNotFoundCondition", got: GatewayNotFoundCondition(gen, msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonGatewayNotFound, wantMsg: msg},
		// VirtualService
		{name: "VirtualServiceConfiguredCondition", got: VirtualServiceConfiguredCondition(gen), wantType: ConditionTypeVirtualService, wantStatus: metav1.ConditionTrue, wantReason: ReasonVirtualServiceReconcileSucceeded},
		{name: "VirtualServiceErrorCondition", got: VirtualServiceErrorCondition(gen, msg), wantType: ConditionTypeVirtualService, wantStatus: metav1.ConditionFalse, wantReason: ReasonVirtualServiceReconcileFailed, wantMsg: msg},
		// RequestAuthentication
		{name: "RequestAuthenticationConfiguredCondition", got: RequestAuthenticationConfiguredCondition(gen), wantType: ConditionTypeRequestAuthentication, wantStatus: metav1.ConditionTrue, wantReason: ReasonRequestAuthenticationReconcileSucceeded},
		{name: "RequestAuthenticationErrorCondition", got: RequestAuthenticationErrorCondition(gen, msg), wantType: ConditionTypeRequestAuthentication, wantStatus: metav1.ConditionFalse, wantReason: ReasonRequestAuthenticationReconcileFailed, wantMsg: msg},
		// AuthorizationPolicy
		{name: "AuthorizationPolicyConfiguredCondition", got: AuthorizationPolicyConfiguredCondition(gen), wantType: ConditionTypeAuthorizationPolicy, wantStatus: metav1.ConditionTrue, wantReason: ReasonAuthorizationPolicyReconcileSucceeded},
		{name: "AuthorizationPolicyErrorCondition", got: AuthorizationPolicyErrorCondition(gen, msg), wantType: ConditionTypeAuthorizationPolicy, wantStatus: metav1.ConditionFalse, wantReason: ReasonAuthorizationPolicyReconcileFailed, wantMsg: msg},
		// Dependencies
		{name: "DependenciesReadyCondition", got: DependenciesReadyCondition(gen), wantType: ConditionTypeDependencies, wantStatus: metav1.ConditionTrue, wantReason: ReasonDependenciesAvailable},
		{name: "DependenciesMissingCondition", got: DependenciesMissingCondition(gen, msg), wantType: ConditionTypeDependencies, wantStatus: metav1.ConditionFalse, wantReason: ReasonDependenciesMissing, wantMsg: msg},
		{name: "DependenciesErrorCondition", got: DependenciesErrorCondition(gen, msg), wantType: ConditionTypeDependencies, wantStatus: metav1.ConditionFalse, wantReason: ReasonDependenciesReconcileFailed, wantMsg: msg},
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
			if tc.wantMsg != "" && tc.got.Message != tc.wantMsg {
				t.Errorf("Message: got %q, want %q", tc.got.Message, tc.wantMsg)
			}
			if tc.got.Message == "" {
				t.Errorf("Message: got empty, want non-empty (required by the condition schema)")
			}
			if tc.got.ObservedGeneration != gen {
				t.Errorf("ObservedGeneration: got %d, want %d", tc.got.ObservedGeneration, gen)
			}
		})
	}
}
