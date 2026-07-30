package vpa

import (
	"testing"

	"github.com/kyma-project/api-gateway/internal/processing"
)

func moduleLabels() map[string]string {
	return map[string]string{
		processing.ModuleLabelKey:       processing.ApiGatewayLabelValue,
		processing.K8sManagedByLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sComponentLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sPartOfLabelKey:    processing.ApiGatewayLabelValue,
	}
}

func TestEnsureModuleLabels_NilInput_AddsAllLabels(t *testing.T) {
	got, changed := ensureModuleLabels(nil)
	if !changed {
		t.Fatal("expected changed=true for nil input")
	}
	assertLabelsContain(t, got, moduleLabels())
}

func TestEnsureModuleLabels_EmptyMap_AddsAllLabels(t *testing.T) {
	got, changed := ensureModuleLabels(map[string]string{})
	if !changed {
		t.Fatal("expected changed=true for empty map")
	}
	assertLabelsContain(t, got, moduleLabels())
}

func TestEnsureModuleLabels_AllModuleLabelsPresent_NoChange(t *testing.T) {
	_, changed := ensureModuleLabels(moduleLabels())

	if changed {
		t.Fatal("expected changed=false when all module labels are already correct")
	}
}

func TestEnsureModuleLabels_ModuleLabelsWrongValues_Corrected(t *testing.T) {
	input := map[string]string{
		processing.ModuleLabelKey:       "wrong-value",
		processing.K8sManagedByLabelKey: "wrong-value",
		processing.K8sComponentLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sPartOfLabelKey:    processing.ApiGatewayLabelValue,
	}

	got, changed := ensureModuleLabels(input)

	if !changed {
		t.Fatal("expected changed=true when module label values are incorrect")
	}
	assertLabelsContain(t, got, moduleLabels())
}

func TestEnsureModuleLabels_DoesNotMutateInputMap(t *testing.T) {
	input := map[string]string{
		processing.ModuleLabelKey: "wrong-value",
		"custom/label":            "custom-value",
	}

	got, changed := ensureModuleLabels(input)

	if !changed {
		t.Fatal("expected changed=true when input is missing or has incorrect module labels")
	}
	if input[processing.ModuleLabelKey] != "wrong-value" {
		t.Fatalf("expected input map to remain unchanged for %s, got %q", processing.ModuleLabelKey, input[processing.ModuleLabelKey])
	}
	if _, exists := input[processing.K8sManagedByLabelKey]; exists {
		t.Fatalf("expected input map to remain unchanged and not gain label %s", processing.K8sManagedByLabelKey)
	}
	assertLabelsContain(t, got, moduleLabels())
	if got["custom/label"] != "custom-value" {
		t.Errorf("expected non-module label custom/label to be preserved, got %q", got["custom/label"])
	}
}

func TestEnsureModuleLabels_NonModuleLabelsOnly_ModuleLabelsAdded(t *testing.T) {
	input := map[string]string{
		"custom/label":  "custom-value",
		"another/label": "another-value",
	}

	got, changed := ensureModuleLabels(input)

	if !changed {
		t.Fatal("expected changed=true when module labels are missing")
	}
	assertLabelsContain(t, got, moduleLabels())
	if got["custom/label"] != "custom-value" {
		t.Errorf("expected non-module label custom/label to be preserved, got %q", got["custom/label"])
	}
	if got["another/label"] != "another-value" {
		t.Errorf("expected non-module label another/label to be preserved, got %q", got["another/label"])
	}
}

func TestEnsureModuleLabels_MixedLabels_CorrectModuleAddsMissing(t *testing.T) {
	input := map[string]string{
		processing.ModuleLabelKey:       processing.ApiGatewayLabelValue,
		processing.K8sManagedByLabelKey: "wrong-value",
		"custom/label":                  "custom-value",
	}

	got, changed := ensureModuleLabels(input)

	if !changed {
		t.Fatal("expected changed=true for mixed label state")
	}
	assertLabelsContain(t, got, moduleLabels())
	if got["custom/label"] != "custom-value" {
		t.Errorf("expected non-module label custom/label to be preserved, got %q", got["custom/label"])
	}
}

func assertLabelsContain(t *testing.T, labels, expected map[string]string) {
	t.Helper()
	for key, value := range expected {
		if labels[key] != value {
			t.Errorf("expected label %s=%s, got %q", key, value, labels[key])
		}
	}
}
