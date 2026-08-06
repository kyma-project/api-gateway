package resourceasserts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

type StructCheck struct {
	Gvk       schema.GroupVersionKind
	Name      string
	Namespace string
}

func AssertResourceExists(t *testing.T, r *resources.Resources, sc StructCheck, checkTimeout time.Duration, _ time.Duration) {
	t.Helper()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(sc.Gvk)
	obj.SetName(sc.Name)
	obj.SetNamespace(sc.Namespace)
	require.NoError(t, wait.For(conditions.New(r).ResourceMatch(obj, func(o k8s.Object) bool { return true }), wait.WithTimeout(checkTimeout)), "%s %s/%s should exist", sc.Gvk.Kind, sc.Namespace, sc.Name)
}

func AssertResourceDoesNotExist(t *testing.T, r *resources.Resources, sc StructCheck, checkTimeout time.Duration, _ time.Duration) {
	t.Helper()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(sc.Gvk)
	obj.SetName(sc.Name)
	obj.SetNamespace(sc.Namespace)
	require.NoError(t, wait.For(conditions.New(r).ResourceDeleted(obj), wait.WithTimeout(checkTimeout)), "%s %s/%s should not exist", sc.Gvk.Kind, sc.Namespace, sc.Name)
}
