package controllers

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type state int

const (
	Ready   state = 0
	Error   state = 1
	Warning state = 2
)

type Status interface {
	NestedError() error
	ToAPIGatewayStatus() (operatorv1alpha1.APIGatewayStatus, error)
	IsReady() bool
	IsWarning() bool
	IsError() bool
}

type status struct {
	err         error
	description string
	state       state
}

func ErrorStatus(err error, description string) Status {

	return status{
		err:         err,
		description: description,
		state:       Error,
	}
}

func WarningStatus(err error, description string) Status {
	return status{
		err:         err,
		description: description,
		state:       Warning,
	}
}

func SuccessfulStatus() Status {
	return status{
		description: "Successfully reconciled",
		state:       Ready,
	}
}

func (s status) NestedError() error {
	return s.err
}

func (s status) ToAPIGatewayStatus() (operatorv1alpha1.APIGatewayStatus, error) {

	switch s.state {
	case Ready:
		return operatorv1alpha1.APIGatewayStatus{
			State:       operatorv1alpha1.Ready,
			Description: "Successfully reconciled",
		}, nil
	case Warning:
		return operatorv1alpha1.APIGatewayStatus{
			State:       operatorv1alpha1.Warning,
			Description: s.description,
		}, nil
	case Error:
		return operatorv1alpha1.APIGatewayStatus{
			State:       operatorv1alpha1.Error,
			Description: s.description,
		}, nil
	default:
		return operatorv1alpha1.APIGatewayStatus{}, fmt.Errorf("unsupported status: %v", s.state)
	}
}

func (s status) IsError() bool {
	return s.state == Error
}

func (s status) IsWarning() bool {
	return s.state == Warning
}

func (s status) IsReady() bool {
	return s.state == Ready
}

func UpdateApiGatewayStatus(ctx context.Context, k8sClient client.Client, apiGatewayCR *operatorv1alpha1.APIGateway, status Status) error {
	newStatus, err := status.ToAPIGatewayStatus()
	if err != nil {
		return err
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := k8sClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); getErr != nil {
			return getErr
		}

		apiGatewayCR.Status = newStatus

		if updateErr := k8sClient.Status().Update(ctx, apiGatewayCR); updateErr != nil {
			return updateErr
		}

		return nil
	})
}
