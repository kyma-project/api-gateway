package v1alpha1

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
		{name: "ErrorCondition/ReconcileFailed", got: ErrorCondition(gen, ReasonReconcileFailed, msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonReconcileFailed, wantMsg: msg},
		{name: "ErrorCondition/ValidationFailed", got: ErrorCondition(gen, ReasonValidationFailed, msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonValidationFailed, wantMsg: msg},
		{name: "ErrorCondition/APIGatewayNotReady", got: ErrorCondition(gen, ReasonAPIGatewayNotReady, msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonAPIGatewayNotReady, wantMsg: msg},
		{name: "ErrorCondition/EnvoyFilterReconcileFailed", got: ErrorCondition(gen, ReasonEnvoyFilterReconcileFailed, msg), wantType: ConditionTypeReady, wantStatus: metav1.ConditionFalse, wantReason: ReasonEnvoyFilterReconcileFailed, wantMsg: msg},
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
