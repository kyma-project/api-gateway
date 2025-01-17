package helpers

import (
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var rateLimitGVR = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "ratelimits"}

func WaitForRateLimit(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, rateLimitName string, retryOpts []retry.Option) error {
	err := retry.Do(func() error {

		status, err := resourceMgr.GetStatus(k8sClient, rateLimitGVR, namespace, rateLimitName)
		if err != nil {
			return err
		}
		state, found, err := unstructured.NestedString(status, "state")
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("conditions not found for the ratelimit: %s", rateLimitName)
		}
		if state != "Ready" {
			return fmt.Errorf("status of RateLimit is %s state instead of ready", state)
		}
		return nil
	}, retryOpts...)

	if err != nil {
		return fmt.Errorf("ratelimit %s is not ready: %w", rateLimitName, err)
	}

	return nil
}
