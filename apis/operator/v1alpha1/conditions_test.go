package v1alpha1

import "testing"

func TestProcessingCondition(t *testing.T) {
	c := ProcessingCondition()

	if c.Type != ConditionTypeReady {
		t.Fatalf("unexpected type: got %q, want %q", c.Type, ConditionTypeReady)
	}
	if c.Reason != ReasonReconcileProcessing {
		t.Fatalf("unexpected reason: got %q, want %q", c.Reason, ReasonReconcileProcessing)
	}
}
