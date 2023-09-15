package controllers

import (
	"context"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusHandler interface {
	UpdateToReady(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) error
	UpdateToError(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway, description string) error
}

type Handler struct {
	client client.Client
}

func NewStatusHandler(client client.Client) Handler {
	return Handler{
		client: client,
	}
}

func (d Handler) update(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) error {
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

func (d Handler) UpdateToReady(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) error {
	apiGatewayCR.Status.State = operatorv1alpha1.Ready
	apiGatewayCR.Status.Description = "Successfully reconciled"
	return d.update(ctx, apiGatewayCR)
}

func (d Handler) UpdateToError(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway, description string) error {
	apiGatewayCR.Status.State = operatorv1alpha1.Error

	apiGatewayCR.Status.Description = description
	return d.update(ctx, apiGatewayCR)
}
