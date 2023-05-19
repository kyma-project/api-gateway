package api_gateway

import (
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"strings"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
)

func SwitchJwtHandler(ts testcontext.Testsuite, jwtHandler string) (string, error) {
	mapper, err := client.GetDiscoveryMapper()
	if err != nil {
		return "", err
	}
	mapping, err := mapper.RESTMapping(schema.ParseGroupKind("ConfigMap"))
	if err != nil {
		return "", err
	}
	currentJwtHandler, configMap, err := getConfigMapJwtHandler(ts.ResourceManager, ts.K8sClient, mapping.Resource)
	if err != nil {
		configMap := unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "ConfigMap",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      testcontext.ConfigMapName,
					"namespace": testcontext.DefaultNS,
				},
				"data": map[string]interface{}{
					"api-gateway-config": "jwtHandler: " + jwtHandler,
				},
			},
		}
		currentJwtHandler = jwtHandler
		err = ts.ResourceManager.CreateResource(ts.K8sClient, mapping.Resource, testcontext.DefaultNS, configMap)
	}
	if err != nil {
		return "", fmt.Errorf("could not get or create jwtHandler config:\n %+v", err)
	}
	if currentJwtHandler != jwtHandler {
		configMap.Object["data"].(map[string]interface{})["api-gateway-config"] = "jwtHandler: " + jwtHandler
		err = ts.ResourceManager.UpdateResource(ts.K8sClient, mapping.Resource, testcontext.DefaultNS, testcontext.ConfigMapName, *configMap)
		if err != nil {
			return "", fmt.Errorf("unable to update ConfigMap:\n %+v", err)
		}
	}
	return currentJwtHandler, err
}

func getConfigMapJwtHandler(resourceManager *resource.Manager, k8sClient dynamic.Interface, gvr schema.GroupVersionResource) (string, *unstructured.Unstructured, error) {
	res, err := resourceManager.GetResource(k8sClient, gvr, testcontext.DefaultNS, testcontext.ConfigMapName)
	if err != nil {
		return "", res, fmt.Errorf("could not get ConfigMap:\n %+v", err)
	}
	data, found, err := unstructured.NestedMap(res.Object, "data")
	if err != nil || !found {
		return "", res, fmt.Errorf("could not find data in the ConfigMap:\n %+v", err)
	}
	return strings.Split(data["api-gateway-config"].(string), ": ")[1], res, nil
}
