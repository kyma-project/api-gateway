package extauth

import (
	"bytes"
	_ "embed"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"

	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
)

//go:embed ext_auth.yaml
var extAuthTemplate []byte

type ExtAuthOptions struct {
	Template []byte
}

type ExtAuthOption func(*ExtAuthOptions)

func WithExtAuthTemplate(template []byte) ExtAuthOption {
	return func(o *ExtAuthOptions) {
		extAuthTemplate = template
	}
}

func CreateExtAuth(t *testing.T) error {
	t.Helper()

	t.Log("Creating external authorizer")
	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
	}

	err = r.Get(t.Context(), "ext-authz", "ext-auth", &unstructured.Unstructured{})
	// If the external authorizer already exists, skip creation
	if err == nil {
		t.Log("External authorizer already exists, skipping creation")
		return nil
	}

	err = infrahelpers.CreateNamespace(t, "ext-auth", infrahelpers.WithLabels(map[string]string{"kubernetes.io/metadata.name": "ext-auth"}), infrahelpers.IgnoreAlreadyExists())
	if err != nil {
		t.Logf("Failed to create namespace for external authorizer: %v", err)
		return err
	}

	err = decoder.DecodeEach(
		t.Context(),
		bytes.NewBuffer(extAuthTemplate),
		decoder.CreateHandler(r),
		decoder.MutateNamespace("ext-auth"),
	)
	if err != nil {
		t.Logf("Failed to create external authorizer: %v", err)
		return err
	}
	setup.DeclareCleanup(t, func() {
		t.Logf("Cleaning up external authorizer")
		err := decoder.DecodeEach(
			setup.GetCleanupContext(),
			bytes.NewBuffer(extAuthTemplate),
			decoder.DeleteHandler(r),
			decoder.MutateNamespace("ext-auth"),
		)
		if err != nil {
			t.Logf("Failed to clean up external authorizer: %v", err)
		} else {
			t.Logf("Successfully cleaned up external authorizer")
		}

	})

	return wait.For(conditions.New(r).DeploymentAvailable("ext-authz", "ext-auth"))
}
