package v2alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVirtualServiceConfiguredCondition(t *testing.T) {
	c := VirtualServiceConfiguredCondition(17)

	if c.Type != ConditionTypeVirtualService {
		t.Fatalf("unexpected type: got %q, want %q", c.Type, ConditionTypeVirtualService)
	}
	if c.Status != metav1.ConditionTrue {
		t.Fatalf("unexpected status: got %q, want %q", c.Status, metav1.ConditionTrue)
	}
	if c.Reason != ReasonVirtualServiceReconcileSucceeded {
		t.Fatalf("unexpected reason: got %q, want %q", c.Reason, ReasonVirtualServiceReconcileSucceeded)
	}
	if c.ObservedGeneration != 17 {
		t.Fatalf("unexpected generation: got %d, want %d", c.ObservedGeneration, 17)
	}
}

func TestReadyCondition(t *testing.T) {
	c := ReadyCondition(21)

	if c.Type != ConditionTypeReady {
		t.Fatalf("unexpected type: got %q, want %q", c.Type, ConditionTypeReady)
	}
	if c.Status != metav1.ConditionTrue {
		t.Fatalf("unexpected status: got %q, want %q", c.Status, metav1.ConditionTrue)
	}
	if c.Reason != ReasonReconcileSucceeded {
		t.Fatalf("unexpected reason: got %q, want %q", c.Reason, ReasonReconcileSucceeded)
	}
	if c.ObservedGeneration != 21 {
		t.Fatalf("unexpected generation: got %d, want %d", c.ObservedGeneration, 21)
	}
}
