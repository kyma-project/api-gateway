package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestConditionTable is a direct machine-readable mirror of the ADR 0015 condition table.
// Each row corresponds to one table entry and asserts (Type, Status, Reason).
func TestConditionTable(t *testing.T) {
	const gen = int64(1)
	const msg = "some message"

	rows := []struct {
		name       string
		got        metav1.Condition
		wantType   string
		wantStatus metav1.ConditionStatus
		wantReason string
	}{
		// DependenciesReady rows
		{"DependenciesReady/True", DependenciesReadyCondition(gen), ConditionTypeDependencies, metav1.ConditionTrue, ReasonDependenciesAvailable},
		{"DependenciesReady/False/Missing", DependenciesMissingCondition(gen, msg), ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesMissing},
		{"DependenciesReady/False/Error", DependenciesErrorCondition(gen, msg), ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesReconcileFailed},
		// Ready rows
		{"Ready/True", ReadyCondition(gen), ConditionTypeReady, metav1.ConditionTrue, ReasonReconcileSucceeded},
		{"Ready/False/APIGatewayNotReady", ErrorCondition(gen, ReasonAPIGatewayNotReady, msg), ConditionTypeReady, metav1.ConditionFalse, ReasonAPIGatewayNotReady},
		{"Ready/False/ValidationFailed", ErrorCondition(gen, ReasonValidationFailed, msg), ConditionTypeReady, metav1.ConditionFalse, ReasonValidationFailed},
		{"Ready/False/EnvoyFilterReconcileFailed", ErrorCondition(gen, ReasonEnvoyFilterReconcileFailed, msg), ConditionTypeReady, metav1.ConditionFalse, ReasonEnvoyFilterReconcileFailed},
		{"Ready/False/ReconcileFailed", ErrorCondition(gen, ReasonReconcileFailed, msg), ConditionTypeReady, metav1.ConditionFalse, ReasonReconcileFailed},
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			if row.got.Type != row.wantType {
				t.Errorf("Type: got %q, want %q", row.got.Type, row.wantType)
			}
			if row.got.Status != row.wantStatus {
				t.Errorf("Status: got %q, want %q", row.got.Status, row.wantStatus)
			}
			if row.got.Reason != row.wantReason {
				t.Errorf("Reason: got %q, want %q", row.got.Reason, row.wantReason)
			}
		})
	}
}

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
