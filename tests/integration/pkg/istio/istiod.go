package istio

import (
	"fmt"
	"log"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// PatchIstiodDeploymentWithEnvironmentVariables patches the istiod deployment with the given environment variables.
func PatchIstiodDeploymentWithEnvironmentVariables(resourceMgr *resource.Manager, k8sClient dynamic.Interface, environmentVariables map[string]string) error {
	log.Printf("Patching istiod deployment with environment variables: %v", environmentVariables)
	res, err := resourceMgr.GetResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, "istio-system", "istiod")
	if err != nil {
		return fmt.Errorf("could not get istiod deployment: %s", err.Error())
	}

	containers, found, err := unstructured.NestedSlice(res.Object, "spec", "template", "spec", "containers")
	if err != nil || !found {
		return fmt.Errorf("could not find containers in istiod deployment: %s", err.Error())
	}
	if len(containers) != 1 {
		return fmt.Errorf("istiod deployment contains more than one container")
	}

	env, found, err := unstructured.NestedSlice(containers[0].(map[string]interface{}), "env")
	if err != nil || !found {
		return fmt.Errorf("could not find env in istiod deployment: %s", err.Error())
	}

	for key, value := range environmentVariables {
		env = append(env, map[string]interface{}{"name": key, "value": value})
	}

	err = unstructured.SetNestedSlice(containers[0].(map[string]interface{}), env, "env")
	if err != nil {
		return fmt.Errorf("could not set env in istiod deployment: %s", err.Error())
	}

	err = unstructured.SetNestedSlice(res.Object, containers, "spec", "template", "spec", "containers")
	if err != nil {
		return fmt.Errorf("could not set containers in istiod deployment: %s", err.Error())
	}

	err = resourceMgr.UpdateResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, "istio-system", "istiod", *res)
	if err != nil {
		return fmt.Errorf("could not update istiod deployment: %s", err.Error())
	}

	return nil
}

// RemoveEnvironmentVariableFromIstiodDeployment removes the given environment variable from the deployment.
func RemoveEnvironmentVariableFromIstiodDeployment(resourceMgr *resource.Manager, k8sClient dynamic.Interface, environmentVariableName string) error {
	log.Printf("Removing environment variable %s from istiod deployment", environmentVariableName)
	res, err := resourceMgr.GetResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, "istio-system", "istiod")
	if err != nil {
		return fmt.Errorf("could not get istiod deployment: %s", err.Error())
	}

	containers, found, err := unstructured.NestedSlice(res.Object, "spec", "template", "spec", "containers")
	if err != nil || !found {
		return fmt.Errorf("could not find containers in istiod deployment: %s", err.Error())
	}
	if len(containers) != 1 {
		return fmt.Errorf("istiod deployment contains more than one container")
	}

	env, found, err := unstructured.NestedSlice(containers[0].(map[string]interface{}), "env")
	if err != nil || !found {
		return fmt.Errorf("could not find env in istiod deployment: %s", err.Error())
	}

	for i, v := range env {
		if v.(map[string]interface{})["name"] == environmentVariableName {
			env = append(env[:i], env[i+1:]...)
			break
		}
	}

	err = unstructured.SetNestedSlice(containers[0].(map[string]interface{}), env, "env")
	if err != nil {
		return fmt.Errorf("could not set env in istiod deployment: %s", err.Error())
	}

	err = unstructured.SetNestedSlice(res.Object, containers, "spec", "template", "spec", "containers")
	if err != nil {
		return fmt.Errorf("could not set containers in istiod deployment: %s", err.Error())
	}

	err = resourceMgr.UpdateResource(k8sClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, "istio-system", "istiod", *res)
	if err != nil {
		return fmt.Errorf("could not update istiod deployment: %s", err.Error())
	}

	return nil
}
