package helpers

import (
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var deploymentGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

func WaitForDeployment(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, deploymentName string, retryOpts []retry.Option) error {
	err := retry.Do(func() error {
		status, err := resourceMgr.GetStatus(k8sClient, deploymentGVR, namespace, deploymentName)
		if err != nil {
			return err
		}
		conditions, found, err := unstructured.NestedSlice(status, "conditions")
		if !found {
			return fmt.Errorf("conditions not found for the deployment: %s", deploymentName)
		}
		if err != nil {
			return err
		}
		for _, condition := range conditions {
			conditionMap := condition.(map[string]interface{})
			conditionReason := conditionMap["reason"].(string)
			conditionStatus := conditionMap["status"].(string)
			if conditionStatus != "True" {
				return fmt.Errorf("status condition %s in the deployment %s is not true", conditionReason, deploymentName)
			}
		}
		return nil
	}, retryOpts...)

	if err != nil {
		return fmt.Errorf("deployment %s is not ready: %w", deploymentName, err)
	}

	return nil
}
