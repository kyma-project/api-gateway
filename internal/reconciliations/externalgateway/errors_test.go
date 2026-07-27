package externalgateway

import (
	"errors"
	"fmt"
	"testing"
)

func TestReasonedError_ErrorAndReason(t *testing.T) {
	err := NewReasonedError("MyReason", "problem with %s", "foo")
	if err.Error() != "problem with foo" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
	reason, ok := ErrorReason(err)
	if !ok || reason != "MyReason" {
		t.Fatalf("expected reason MyReason, got %q, ok=%v", reason, ok)
	}
}

func TestReasonedError_WrapPreservesReason(t *testing.T) {
	inner := NewReasonedError("InnerReason", "boom")
	wrapped := fmt.Errorf("context: %w", inner)
	reason, ok := ErrorReason(wrapped)
	if !ok || reason != "InnerReason" {
		t.Fatalf("expected InnerReason through wrap, got %q, ok=%v", reason, ok)
	}
}

func TestAsReasoned_PlainErrorReturnsFalse(t *testing.T) {
	_, ok := ErrorReason(errors.New("plain"))
	if ok {
		t.Fatal("expected false for plain error")
	}
}
