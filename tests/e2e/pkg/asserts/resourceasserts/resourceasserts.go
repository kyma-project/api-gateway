package resourceasserts

import (
	"fmt"
	"slices"
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

const ManagedByDisclaimerAnnotation = "DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."

type StructCheck struct {
	Gvk       schema.GroupVersionKind
	Name      string
	Namespace string
}

func Resource(group, version, kind, name, namespace string) StructCheck {
	return StructCheck{
		Gvk:       schema.GroupVersionKind{Group: group, Version: version, Kind: kind},
		Name:      name,
		Namespace: namespace,
	}
}

func AssertResourceExists(t *testing.T, r *resources.Resources, sc StructCheck, checkTimeout time.Duration) {
	t.Helper()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(sc.Gvk)
	obj.SetName(sc.Name)
	obj.SetNamespace(sc.Namespace)
	require.NoError(t, wait.For(conditions.New(r).ResourceMatch(obj, func(o k8s.Object) bool { return true }), wait.WithTimeout(checkTimeout)), "%s %s/%s should exist", sc.Gvk.Kind, sc.Namespace, sc.Name)

	fetched := &unstructured.Unstructured{}
	fetched.SetGroupVersionKind(sc.Gvk)
	require.NoError(t, r.Get(t.Context(), sc.Name, sc.Namespace, fetched), "fetching %s %s/%s for annotation/label check", sc.Gvk.Kind, sc.Namespace, sc.Name)

	err := checkModuleAnnotationsAndLabels(fetched)
	require.NoError(t, err)
}

func AssertResourceDoesNotExist(t *testing.T, r *resources.Resources, sc StructCheck, checkTimeout time.Duration) {
	t.Helper()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(sc.Gvk)
	obj.SetName(sc.Name)
	obj.SetNamespace(sc.Namespace)
	require.NoError(t, wait.For(conditions.New(r).ResourceDeleted(obj), wait.WithTimeout(checkTimeout)), "%s %s/%s should not exist", sc.Gvk.Kind, sc.Namespace, sc.Name)
}

func checkModuleAnnotationsAndLabels(obj *unstructured.Unstructured) error {

	if isManagedByGardener(obj) {
		return nil
	}
	if isOperatorResource(obj) {
		requiredOperatorResourceLabels := []string{"app.kubernetes.io/name", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component", "app.kubernetes.io/part-of", "kyma-project.io/module"}

		labels := obj.GetLabels()
		for _, label := range requiredOperatorResourceLabels {
			if _, found := labels[label]; !found {
				return fmt.Errorf("kind: %s, name: %s, does not contain required label: %s", obj.GetKind(), obj.GetName(), label)
			}
		}
	} else {
		// Verify resources of external components like Oathkeeper
		annotations := obj.GetAnnotations()
		if annotations["apigateways.operator.kyma-project.io/managed-by-disclaimer"] != ManagedByDisclaimerAnnotation {
			return fmt.Errorf("kind: %s, name: %s, does not have required annotation disclaimer", obj.GetKind(), obj.GetName())
		}

		labels := obj.GetLabels()
		moduleLabel := "kyma-project.io/module"
		if _, found := labels[moduleLabel]; !found {
			return fmt.Errorf("kind: %s, name: %s, does not contain required label: %s", obj.GetKind(), obj.GetName(), moduleLabel)
		}
	}

	return nil
}

type operatorResourceKey struct {
	kind string
	name string
}

func isOperatorResource(obj *unstructured.Unstructured) bool {
	operatorResources := []operatorResourceKey{
		{kind: "CustomResourceDefinition", name: "apigateways.operator.kyma-project.io"},
		{kind: "CustomResourceDefinition", name: "apirules.gateway.kyma-project.io"},
		{kind: "Deployment", name: "api-gateway-controller-manager"},
		{kind: "Service", name: "api-gateway-operator-metrics"},
		{kind: "ClusterRole", name: "api-gateway-manager-role"},
		{kind: "ClusterRoleBinding", name: "api-gateway-manager-rolebinding"},
		{kind: "Role", name: "api-gateway-leader-election-role"},
		{kind: "RoleBinding", name: "api-gateway-leader-election-rolebinding"},
		{kind: "Gateway", name: "kyma-gateway"},
		{kind: "Certificate", name: "kyma-tls-cert"},
		{kind: "VirtualService", name: "istio-healthz"},
		{kind: "ConfigMap", name: "api-gateway-apirule-ui.operator.kyma-project.io"},
		{kind: "ConfigMap", name: "api-gateway-ui.operator.kyma-project.io"},
		{kind: "PriorityClass", name: "api-gateway-priority-class"},
		{kind: "ServiceAccount", name: "api-gateway-controller-manager"},
	}

	key := operatorResourceKey{kind: obj.GetKind(), name: obj.GetName()}
	return slices.Contains(operatorResources, key)
}

func isManagedByGardener(obj *unstructured.Unstructured) bool {
	gardenerResources := []string{"kyma-gateway-certs"}
	return slices.Contains(gardenerResources, obj.GetName())
}
