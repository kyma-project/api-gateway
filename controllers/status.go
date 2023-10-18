package controllers

import (
	"context"
	"fmt"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type State int

const (
	Ready      State = 0
	Error      State = 1
	Warning    State = 2
	Deleting   State = 3
	Processing State = 4
)

type Status interface {
	NestedError() error
	ToAPIGatewayStatus() (operatorv1alpha1.APIGatewayStatus, error)
	IsReady() bool
	IsWarning() bool
	IsError() bool
	State() State
	Description() string
}

type status struct {
	err         error
	description string
	state       State
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

func ReadyStatus() Status {
	return status{
		description: "Successfully reconciled",
		state:       Ready,
	}
}

func DeletingStatus() Status {
	return status{
		state: Deleting,
	}
}

func ProcessingStatus() Status {
	return status{
		state: Processing,
	}
}

func (s status) NestedError() error {
	return s.err
}

func (s status) Description() string {
	return s.description
}

func (s status) ToAPIGatewayStatus() (operatorv1alpha1.APIGatewayStatus, error) {
	switch s.state {
	case Ready:
		return operatorv1alpha1.APIGatewayStatus{
			State:       operatorv1alpha1.Ready,
			Description: s.description,
		}, nil
	case Processing:
		return operatorv1alpha1.APIGatewayStatus{
			State:       operatorv1alpha1.Processing,
			Description: s.description,
		}, nil
	case Warning:
		return operatorv1alpha1.APIGatewayStatus{
			State:       operatorv1alpha1.Warning,
			Description: s.description,
		}, nil
	case Deleting:
		return operatorv1alpha1.APIGatewayStatus{
			State:       operatorv1alpha1.Deleting,
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

func (s status) State() State {
	return s.state
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
