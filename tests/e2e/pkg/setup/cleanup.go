package setup

import (
	"context"
	"os"
	"testing"
)

var forceSkipCleanup = os.Getenv("FORCE_SKIP_CLEANUP") == "true"
var shouldSkipCleanup = os.Getenv("SKIP_CLEANUP") == "true"

func ShouldSkipCleanup(t *testing.T) bool {
	if forceSkipCleanup {
		t.Logf("FORCE_SKIP_CLEANUP is set, skipping cleanup")
		return true
	}
	return t.Failed() && shouldSkipCleanup
}

func DeclareCleanup(t *testing.T, f func()) {
	t.Helper()
	t.Cleanup(func() {
		t.Helper()
		if ShouldSkipCleanup(t) {
			t.Logf("Either tests failed or FORCE_SKIP_CLEANUP is set; skipping test cleanup")
			return
		}
		t.Logf("Cleaning up")
		f()
	})
}

func GetCleanupContext() context.Context {
	return context.Background()
}
