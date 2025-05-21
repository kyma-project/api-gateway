package gateway

import (
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func checkModuleAnnotationsAndLabels(obj *unstructured.Unstructured) error {
	if isManagedByGardener(obj) {
		return nil
	}

	if isOperatorResource(obj) {
		requiredOperatorResourceLabels := []string{
			"app.kubernetes.io/name",
			"app.kubernetes.io/instance",
			"app.kubernetes.io/version",
			"app.kubernetes.io/component",
			"app.kubernetes.io/part-of",
			"kyma-project.io/module",
		}

		labels := obj.GetLabels()
		for _, label := range requiredOperatorResourceLabels {
			if _, found := labels[label]; !found {
				return fmt.Errorf("kind: %s, name: %s, does not contain required label: %s", obj.GetKind(), obj.GetName(), label)
			}
		}
	} else {
		// Verify resources of external components like Oathkeeper
		annotations := obj.GetAnnotations()
		if annotations["apigateways.operator.kyma-project.io/managed-by-disclaimer"] != "DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state." {
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

func isOperatorResource(obj *unstructured.Unstructured) bool {
	operatorResources := []string{"apigateways.operator.kyma-project.io", "apirules.gateway.kyma-project.io", "api-gateway-controller-manager", "api-gateway-operator-metrics",
		"api-gateway-manager-role", "api-gateway-manager-rolebinding", "api-gateway-leader-election-role", "api-gateway-leader-election-rolebinding", "kyma-gateway", "kyma-tls-cert",
		"istio-healthz", "api-gateway-apirule-ui.operator.kyma-project.io", "api-gateway-ui.operator.kyma-project.io", "api-gateway-priority-class"}

	return slices.Contains(operatorResources, obj.GetName())
}

func isManagedByGardener(obj *unstructured.Unstructured) bool {
	gardenerResources := []string{"kyma-gateway-certs"}
	return slices.Contains(gardenerResources, obj.GetName())
}
