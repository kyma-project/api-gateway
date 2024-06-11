package hooks

import (
	"context"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var VerifyIfControllerHasBeenRestarted = func(ctx context.Context, sc *godog.Scenario, _ error) (context.Context, error) {
	c := k8sclient.GetK8sClient()

	podList := &corev1.PodList{}
	err := c.List(ctx, podList, client.MatchingLabels{"app.kubernetes.io/component": "api-gateway-operator.kyma-project.io"})
	if err != nil {
		return ctx, err
	}
	if len(podList.Items) < 1 {
		return ctx, errors.New("Controller not found")
	}

	for _, cpod := range podList.Items {
		if len(cpod.Status.ContainerStatuses) > 0 {
			if rc := cpod.Status.ContainerStatuses[0].RestartCount; rc > 0 {
				errMsg := fmt.Sprintf("Controller has been restarted %d times", rc)
				return ctx, errors.New(errMsg)
			}
		}
	}

	return ctx, nil
}
