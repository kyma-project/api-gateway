package operator

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type status interface {
	updateToReady(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) error
}

type StatusHandler struct {
	client client.Client
}

func newStatusHandler(client client.Client) StatusHandler {
	return StatusHandler{
		client: client,
	}
}

func (d StatusHandler) update(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) error {
	newStatus := apiGatewayCR.Status
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := d.client.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); getErr != nil {
			return getErr
		}

		apiGatewayCR.Status = newStatus

		if updateErr := d.client.Status().Update(ctx, apiGatewayCR); updateErr != nil {
			return updateErr
		}

		return nil
	})
}

func (d StatusHandler) updateToReady(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) error {
	apiGatewayCR.Status.State = operatorv1alpha1.Ready
	apiGatewayCR.Status.Description = fmt.Sprintf("Successfully reconciled")
	return d.update(ctx, apiGatewayCR)
}
