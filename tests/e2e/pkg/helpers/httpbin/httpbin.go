package httpbin

import (
	"bytes"
	_ "embed"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
)

//go:embed manifest.yaml
var manifest []byte

//go:embed manifest_second.yaml
var manifestSecondHttpbin []byte

func DeployHttpbin(t *testing.T, namespace string) (svcName string, svcPort int, err error) {
	t.Helper()

	httpbinName, err := getNameFromManifest(manifest)
	if err != nil {
		t.Logf("Error getting name from Manifest: %v", err)
	}

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return "", 0, fmt.Errorf("failed to get resources client: %w", err)
	}

	return httpbinName, 8000, start(t, r, manifest, httpbinName, namespace)
}

func DeploySecondHttpbin(t *testing.T, namespace string) (svcName string, svcPort int, err error) {
	t.Helper()

	httpbinName, err := getNameFromManifest(manifestSecondHttpbin)
	if err != nil {
		t.Logf("Error getting name from Manifest: %v", err)
	}

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return "", 0, fmt.Errorf("failed to get resources client: %w", err)
	}

	return httpbinName, 8000, start(t, r, manifestSecondHttpbin, httpbinName, namespace)
}

func start(t *testing.T, r *resources.Resources, manifest []byte, name, namespace string) error {
	err := decoder.DecodeEach(
		t.Context(),
		bytes.NewBuffer(manifest),
		decoder.CreateHandler(r),
		decoder.MutateNamespace(namespace),
		decoder.MutateOption(mutateServiceIPFamily),
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

	return wait.For(conditions.New(r).DeploymentAvailable(name, namespace))
}

// mutateServiceIPFamily aligns every core/v1 Service in the deployed
// bundle with the TEST_IP_FAMILY axis via ipfamily.ApplyToService. Under
// TEST_IP_FAMILY=dualstack the Service becomes PreferDualStack +
// [IPv6, IPv4]; single-family modes get the equivalent SingleStack
// setting. Unset TEST_IP_FAMILY leaves the Service on SingleStack IPv4,
// matching pre-dualstack k3d behaviour byte for byte. Non-Service
// objects pass through unchanged.
func mutateServiceIPFamily(obj k8s.Object) error {
	if svc, ok := obj.(*corev1.Service); ok {
		ipfamily.ApplyToService(svc)
	}
	return nil
}

func getNameFromManifest(manifest []byte) (string, error) {
	var m struct {
		Metadata struct {
			Name string `json:"name" yaml:"name"`
		} `json:"metadata" yaml:"metadata"`
	}

	if err := yaml.Unmarshal(manifest, &m); err != nil {
		return "", err
	}
	return m.Metadata.Name, nil
}
