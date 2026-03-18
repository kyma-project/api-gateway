package vpa_test

import (
	"context"
	"testing"

	"github.com/kyma-project/api-gateway/internal/vpa"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var vpaGVK = schema.GroupVersionKind{
	Group:   "autoscaling.k8s.io",
	Version: "v1",
	Kind:    "VerticalPodAutoscaler",
}

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
	s.AddKnownTypeWithName(vpaGVK, &unstructured.Unstructured{})
	return s
}

func existingVPA() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(vpaGVK)
	obj.SetName("api-gateway-controller-manager-vpa")
	obj.SetNamespace("kyma-system")
	obj.Object["spec"] = map[string]interface{}{
		"targetRef": map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"name":       "old-deployment",
		},
	}
	return obj
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

	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(vpaGVK)
	err = c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got)
	if err != nil {
		t.Fatalf("expected VPA to be created, got error: %v", err)
	}

	spec, ok := got.Object["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected spec to be present")
	}
	targetRef, ok := spec["targetRef"].(map[string]interface{})
	if !ok {
		t.Fatal("expected targetRef in spec")
	}
	if targetRef["name"] != "api-gateway-controller-manager" {
		t.Errorf("expected targetRef name to be api-gateway-controller-manager, got %v", targetRef["name"])
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

	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(vpaGVK)
	err = c.Get(context.Background(), types.NamespacedName{Name: "api-gateway-controller-manager-vpa", Namespace: "kyma-system"}, got)
	if err != nil {
		t.Fatalf("expected VPA to exist, got error: %v", err)
	}

	spec := got.Object["spec"].(map[string]interface{})
	targetRef := spec["targetRef"].(map[string]interface{})
	if targetRef["name"] != "api-gateway-controller-manager" {
		t.Errorf("expected targetRef to be updated to api-gateway-controller-manager, got %v", targetRef["name"])
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

	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(vpaGVK)
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
