package artifacts

import (
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultArtifactsDir    = "test-artifacts"
	testArtifactsDirEnvVar = "TEST_ARTIFACTS_DIR"
)

var (
	testRunTimestamp string
	timestampOnce    sync.Once
)

// Root returns the root directory for test artifacts.
// It reads TEST_ARTIFACTS_DIR from the environment, defaulting to ./test-artifacts.
func Root() string {
	if dir, ok := os.LookupEnv(testArtifactsDirEnvVar); ok {
		return dir
	}
	return defaultArtifactsDir
}

// TestRunTimestamp returns a stable timestamp for the current test run.
func TestRunTimestamp() string {
	timestampOnce.Do(func() {
		testRunTimestamp = time.Now().Format("02_01_2006-15_04_05CET")
	})
	return testRunTimestamp
}

// SanitizePathComponent replaces characters that are unsafe in file/directory names.
func SanitizePathComponent(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_", " ", "_", "(", "", ")", "", ",", "",
	)
	return replacer.Replace(name)
}
