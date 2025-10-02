package setup

import (
	"context"
	"os"
	"testing"
)

var shouldSkipCleanup = os.Getenv("SKIP_CLEANUP") == "true"

func ShouldSkipCleanup(t *testing.T) bool {
	return t.Failed() && shouldSkipCleanup
}

func DeclareCleanup(t *testing.T, f func()) {
	t.Helper()
	t.Cleanup(func() {
		t.Helper()
		DumpClusterResources(t)
		if ShouldSkipCleanup(t) {
			t.Logf("Skipping cleanup due to test failure and SKIP_CLEANUP environment variable set to true")
			return
		}
		t.Logf("Cleaning up")
		f()
	})
}

func GetCleanupContext() context.Context {
	return context.Background()
}
