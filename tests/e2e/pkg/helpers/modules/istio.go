package modules

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

//go:embed operator_v1alpha2_istio_ext_authorizers.yaml
var IstioExtAuthorizersTemplate string

// PatchIstioCR reads the current Istio CR, applies the given template as its new state,
// and registers a cleanup that restores the original.
func PatchIstioCR(t *testing.T, template string) error {
	t.Helper()
	t.Log("Patching Istio custom resource")

	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	// Read the current CR so we can restore it in cleanup
	originalCR := &unstructured.Unstructured{}
	originalCR.SetGroupVersionKind(istioGVK())
	if err := r.Get(t.Context(), "default", "kyma-system", originalCR); err != nil {
		t.Logf("Failed to read current Istio custom resource: %v", err)
		return err
	}

	desired := &unstructured.Unstructured{}
	if err := decoder.Decode(bytes.NewBufferString(template), desired); err != nil {
		t.Logf("Failed to decode Istio custom resource template: %v", err)
		return err
	}
	desired.SetResourceVersion(originalCR.GetResourceVersion())

	if err := r.Update(t.Context(), desired); err != nil {
		t.Logf("Failed to patch Istio custom resource: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Log("Restoring Istio custom resource to original state")
		current := &unstructured.Unstructured{}
		current.SetGroupVersionKind(istioGVK())
		if err := r.Get(setup.GetCleanupContext(), "default", "kyma-system", current); err != nil {
			t.Logf("Failed to read Istio custom resource for restore: %v", err)
			return
		}
		originalCR.SetResourceVersion(current.GetResourceVersion())
		if err := r.Update(setup.GetCleanupContext(), originalCR); err != nil {
			t.Logf("Failed to restore Istio custom resource: %v", err)
		} else {
			t.Log("Istio custom resource restored successfully")
		}
	})

	return nil
}

func istioGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "operator.kyma-project.io",
		Version: "v1alpha2",
		Kind:    "Istio",
	}
}
