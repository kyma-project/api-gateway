package vpa_test

import (
	"context"
	"testing"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/vpa"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func vpaCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "verticalpodautoscalers.autoscaling.k8s.io",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "autoscaling.k8s.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "verticalpodautoscalers",
				Singular: "verticalpodautoscaler",
				Kind:     "VerticalPodAutoscaler",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: "v1", Served: true, Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type:                   "object",
							XPreserveUnknownFields: boolPtr(true),
						},
					},
				},
			},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = apiextensionsv1.AddToScheme(s)
	_ = vpav1.AddToScheme(s)
	return s
}

func existingVPA() *vpav1.VerticalPodAutoscaler {
	return &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-gateway-controller-manager-vpa",
			Namespace: "kyma-system",
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef: &autoscaling.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "old-deployment",
			},
		},
	}
}

func existingVPACheckpoint(labels map[string]string) *vpav1.VerticalPodAutoscalerCheckpoint {
	return &vpav1.VerticalPodAutoscalerCheckpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-gateway-controller-manager-vpa-manager",
			Namespace: "kyma-system",
			Labels:    labels,
		},
	}
}

func expectedModuleLabels() map[string]string {
	return map[string]string{
		processing.ModuleLabelKey:       processing.ApiGatewayLabelValue,
		processing.K8sManagedByLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sComponentLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sPartOfLabelKey:    processing.ApiGatewayLabelValue,
	}
}

func TestReconcile_NoCRD_Skips(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	r := vpa.NewReconciler(c)
	err := r.Reconcile(context.Background(), false)
	if err != nil {
		t.Fatalf("expected no error when VPA CRD is not installed, got: %v", err)
	}
}

func TestReconcile_CRDInstalled_CreatesVPA(t *testing.T) {
	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD()).
		Build()

	r := vpa.NewReconciler(c)
	err := r.Reconcile(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &vpav1.VerticalPodAutoscaler{}
	err = c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got)
	if err != nil {
		t.Fatalf("expected VPA to be created, got error: %v", err)
	}

	if got.Spec.TargetRef == nil {
		t.Fatal("expected targetRef to be present")
	}
	if got.Spec.TargetRef.Name != "api-gateway-controller-manager" {
		t.Errorf("expected targetRef name to be api-gateway-controller-manager, got %v", got.Spec.TargetRef.Name)
	}
}

func TestReconcile_CRDInstalled_UpdatesExistingVPA(t *testing.T) {
	existing := existingVPA()

	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD(), existing).
		Build()

	r := vpa.NewReconciler(c)
	err := r.Reconcile(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &vpav1.VerticalPodAutoscaler{}
	err = c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got)
	if err != nil {
		t.Fatalf("expected VPA to exist, got error: %v", err)
	}

	if got.Spec.TargetRef.Name != "api-gateway-controller-manager" {
		t.Errorf("expected targetRef to be updated to api-gateway-controller-manager, got %v", got.Spec.TargetRef.Name)
	}
}

func TestReconcile_CRDInstalled_SkipsUpdateWhenSpecUnchanged(t *testing.T) {
	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD()).
		Build()

	r := vpa.NewReconciler(c)

	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error on first reconcile: %v", err)
	}

	got := &vpav1.VerticalPodAutoscaler{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got); err != nil {
		t.Fatalf("expected VPA to exist: %v", err)
	}
	rvBefore := got.ResourceVersion

	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error on second reconcile: %v", err)
	}

	if err := c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got); err != nil {
		t.Fatalf("expected VPA to exist: %v", err)
	}
	if got.ResourceVersion != rvBefore {
		t.Errorf("expected resource version to remain %s, got %s (unnecessary update)", rvBefore, got.ResourceVersion)
	}
}

func TestReconcile_Deletion_DeletesVPA(t *testing.T) {
	existing := existingVPA()

	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD(), existing).
		Build()

	r := vpa.NewReconciler(c)
	err := r.Reconcile(context.Background(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &vpav1.VerticalPodAutoscaler{}
	err = c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got)
	if err == nil {
		t.Fatal("expected VPA to be deleted, but it still exists")
	}
}

func TestReconcile_Deletion_NoCRD_Skips(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	r := vpa.NewReconciler(c)
	err := r.Reconcile(context.Background(), true)
	if err != nil {
		t.Fatalf("expected no error when VPA CRD is not installed during deletion, got: %v", err)
	}
}

func TestReconcile_Deletion_NoExistingVPA_Succeeds(t *testing.T) {
	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD()).
		Build()

	r := vpa.NewReconciler(c)
	err := r.Reconcile(context.Background(), true)
	if err != nil {
		t.Fatalf("expected no error when deleting non-existent VPA, got: %v", err)
	}
}

func TestReconcile_CRDInstalled_CheckpointWithoutLabels_PatchesLabels(t *testing.T) {
	checkpoint := existingVPACheckpoint(nil)
	vpaInstance := existingVPA()

	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD(), vpaInstance, checkpoint).
		Build()

	r := vpa.NewReconciler(c)

	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &vpav1.VerticalPodAutoscalerCheckpoint{}
	err := c.Get(context.Background(), types.NamespacedName{Name: checkpoint.Name, Namespace: checkpoint.Namespace}, got)
	if err != nil {
		t.Fatalf("expected checkpoint to exist, got error: %v", err)
	}

	expected := expectedModuleLabels()
	for key, value := range expected {
		if got.Labels[key] != value {
			t.Fatalf("expected label %s=%s, got %q", key, value, got.Labels[key])
		}
	}
}

func TestReconcile_CRDInstalled_CheckpointWithLabels_Idempotent(t *testing.T) {
	checkpoint := existingVPACheckpoint(expectedModuleLabels())
	vpaInstance := existingVPA()

	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD(), vpaInstance, checkpoint).
		Build()

	r := vpa.NewReconciler(c)

	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error on first reconcile: %v", err)
	}

	got := &vpav1.VerticalPodAutoscalerCheckpoint{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: checkpoint.Name, Namespace: checkpoint.Namespace}, got); err != nil {
		t.Fatalf("expected checkpoint to exist: %v", err)
	}
	rvBefore := got.ResourceVersion

	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error on second reconcile: %v", err)
	}

	if err := c.Get(context.Background(), types.NamespacedName{Name: checkpoint.Name, Namespace: checkpoint.Namespace}, got); err != nil {
		t.Fatalf("expected checkpoint to exist: %v", err)
	}
	if got.ResourceVersion != rvBefore {
		t.Fatalf("expected checkpoint resource version to remain %s, got %s", rvBefore, got.ResourceVersion)
	}
}

func TestReconcile_CRDInstalled_CheckpointWithNonModuleLabels_PreservesThemAfterPatch(t *testing.T) {
	checkpoint := existingVPACheckpoint(map[string]string{
		"custom/label": "custom-value",
	})
	vpaInstance := existingVPA()

	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD(), vpaInstance, checkpoint).
		Build()

	r := vpa.NewReconciler(c)
	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &vpav1.VerticalPodAutoscalerCheckpoint{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: checkpoint.Name, Namespace: checkpoint.Namespace}, got); err != nil {
		t.Fatalf("expected checkpoint to exist: %v", err)
	}

	for key, value := range expectedModuleLabels() {
		if got.Labels[key] != value {
			t.Errorf("expected module label %s=%s, got %q", key, value, got.Labels[key])
		}
	}
	if got.Labels["custom/label"] != "custom-value" {
		t.Errorf("expected non-module label custom/label to be preserved, got %q", got.Labels["custom/label"])
	}
}

func TestReconcile_CRDInstalled_CheckpointWithWrongModuleLabels_CorrectsThem(t *testing.T) {
	stale := map[string]string{
		processing.ModuleLabelKey:       "wrong-value",
		processing.K8sManagedByLabelKey: "wrong-value",
		processing.K8sComponentLabelKey: processing.ApiGatewayLabelValue,
		processing.K8sPartOfLabelKey:    processing.ApiGatewayLabelValue,
	}
	checkpoint := existingVPACheckpoint(stale)
	vpaInstance := existingVPA()

	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD(), vpaInstance, checkpoint).
		Build()

	r := vpa.NewReconciler(c)
	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &vpav1.VerticalPodAutoscalerCheckpoint{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: checkpoint.Name, Namespace: checkpoint.Namespace}, got); err != nil {
		t.Fatalf("expected checkpoint to exist: %v", err)
	}

	for key, value := range expectedModuleLabels() {
		if got.Labels[key] != value {
			t.Errorf("expected corrected label %s=%s, got %q", key, value, got.Labels[key])
		}
	}
}

func TestReconcile_CRDInstalled_CreatedVPAHasModuleLabels(t *testing.T) {
	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD()).
		Build()

	r := vpa.NewReconciler(c)
	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := &vpav1.VerticalPodAutoscaler{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got); err != nil {
		t.Fatalf("expected VPA to exist: %v", err)
	}

	for key, value := range expectedModuleLabels() {
		if got.Labels[key] != value {
			t.Errorf("expected VPA label %s=%s, got %q", key, value, got.Labels[key])
		}
	}
}

func TestReconcile_CRDInstalled_CheckpointMissing_NoError(t *testing.T) {
	vpaInstance := existingVPA()

	c := fake.NewClientBuilder().
		WithScheme(newScheme()).
		WithObjects(vpaCRD(), vpaInstance).
		Build()

	r := vpa.NewReconciler(c)
	if err := r.Reconcile(context.Background(), false); err != nil {
		t.Fatalf("expected no error when checkpoint is absent, got: %v", err)
	}
}
