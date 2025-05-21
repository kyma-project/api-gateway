package api_gateway

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

const (
	configMapNs   = "kyma-system"
	configMapName = "api-gateway-config"
)

func SwitchJwtHandler(ts testcontext.Testsuite, jwtHandler string) (string, error) {
	log.Printf("Switching JWT handler to %s", jwtHandler)
	mapper, err := client.GetDiscoveryMapper()
	if err != nil {
		return "", err
	}
	mapping, err := mapper.RESTMapping(schema.ParseGroupKind("ConfigMap"))
	if err != nil {
		return "", err
	}
	currentJwtHandler, configMap, err := getConfigMapJwtHandler(ts.ResourceManager(), ts.K8sClient(), mapping.Resource)
	if err != nil {
		configMap := unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "ConfigMap",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"name":      configMapName,
					"namespace": configMapNs,
				},
				"data": map[string]interface{}{
					"api-gateway-config": "jwtHandler: " + jwtHandler,
				},
			},
		}
		currentJwtHandler = jwtHandler
		err = ts.ResourceManager().CreateResource(ts.K8sClient(), mapping.Resource, configMapNs, configMap)
	}
	if err != nil {
		return "", fmt.Errorf("could not get or create jwtHandler config:\n %+v", err)
	}
	if currentJwtHandler != jwtHandler {
		configMap.Object["data"].(map[string]interface{})["api-gateway-config"] = "jwtHandler: " + jwtHandler
		err = ts.ResourceManager().UpdateResource(ts.K8sClient(), mapping.Resource, configMapNs, configMapName, *configMap)
		if err != nil {
			return "", fmt.Errorf("unable to update ConfigMap:\n %+v", err)
		}
	}
	return currentJwtHandler, err
}

func getConfigMapJwtHandler(resourceManager *resource.Manager, k8sClient dynamic.Interface, gvr schema.GroupVersionResource) (string, *unstructured.Unstructured, error) {
	// We want to fail fast in the event of a missing ConfigMap, as this will happen with every initial run of the integration tests because the ConfigMap does not yet exist.
	retryOpts := []retry.Option{
		retry.Delay(time.Duration(1) * time.Second),
		retry.Attempts(3),
		retry.DelayType(retry.FixedDelay),
	}

	res, err := resourceManager.GetResource(k8sClient, gvr, configMapNs, configMapName, retryOpts...)
	if err != nil {
		return "", res, fmt.Errorf("could not get ConfigMap:\n %+v", err)
	}
	data, found, err := unstructured.NestedMap(res.Object, "data")
	if err != nil || !found {
		return "", res, fmt.Errorf("could not find data in the ConfigMap:\n %+v", err)
	}
	return strings.Split(data["api-gateway-config"].(string), ": ")[1], res, nil
}
