package httpbin

import (
	"bytes"
	_ "embed"
	"fmt"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"testing"
)

//go:embed manifest.yaml
var manifest []byte

func DeployHttpbin(t *testing.T, namespace string) (svcName string, svcPort int, err error) {
	t.Helper()

	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return "", 0, fmt.Errorf("failed to get resources client: %w", err)
	}

	return "httpbin", 8000, start(t, r, namespace)
}

func start(t *testing.T, r *resources.Resources, namespace string) error {
	err := decoder.DecodeEach(
		t.Context(),
		bytes.NewBuffer(manifest),
		decoder.CreateHandler(r),
		decoder.MutateNamespace(namespace),
	)
	if err != nil {
		t.Logf("Failed to deploy mock: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Cleaning up httpbin in namespace %s", namespace)
		err := decoder.DecodeEach(
			setup.GetCleanupContext(),
			bytes.NewBuffer(manifest),
			decoder.DeleteHandler(r),
			decoder.MutateNamespace(namespace),
		)
		if err != nil {
			t.Logf("Failed to clean up httpbin: %v", err)
		} else {
			t.Logf("Successfully cleaned up httpbin in namespace %s", namespace)
		}
	})

	return wait.For(conditions.New(r).DeploymentAvailable("httpbin", namespace))
}
