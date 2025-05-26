package controllers

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	processingStatus "github.com/kyma-project/api-gateway/internal/processing/status"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	V2alpha1Status() (processingStatus.ReconciliationV2alpha1Status, error)
	IsReady() bool
	IsWarning() bool
	IsError() bool
	State() State
	Description() string
	Condition() *metav1.Condition
}

type status struct {
	err         error
	description string
	state       State
	condition   *metav1.Condition
}

func ErrorStatus(err error, description string, condition *metav1.Condition) Status {

	return status{
		err:         err,
		description: description,
		state:       Error,
		condition:   condition,
	}
}

func WarningStatus(err error, description string, condition *metav1.Condition) Status {
	return status{
		err:         err,
		description: description,
		state:       Warning,
		condition:   condition,
	}
}

func ReadyStatus(condition *metav1.Condition) Status {
	return status{
		description: "Successfully reconciled",
		state:       Ready,
		condition:   condition,
	}
}

func DeletingStatus(condition *metav1.Condition) Status {
	return status{
		state:     Deleting,
		condition: condition,
	}
}

func ProcessingStatus(condition *metav1.Condition) Status {
	return status{
		state:     Processing,
		condition: condition,
	}
}

func (s status) NestedError() error {
	return s.err
}

func (s status) Description() string {
	return s.description
}

func (s status) ToAPIGatewayStatus() (operatorv1alpha1.APIGatewayStatus, error) {
	newStatus := operatorv1alpha1.APIGatewayStatus{
		Description: s.description,
	}
	if s.condition != nil {
		meta.SetStatusCondition(&newStatus.Conditions, *s.condition)
	}
	switch s.state {
	case Ready:
		newStatus.State = operatorv1alpha1.Ready
		return newStatus, nil
	case Processing:
		newStatus.State = operatorv1alpha1.Processing
		return newStatus, nil
	case Warning:
		newStatus.State = operatorv1alpha1.Warning
		return newStatus, nil
	case Deleting:
		newStatus.State = operatorv1alpha1.Deleting
		return newStatus, nil
	case Error:
		newStatus.State = operatorv1alpha1.Error
		return newStatus, nil
	default:
		return operatorv1alpha1.APIGatewayStatus{}, fmt.Errorf("unsupported status state: %v", s.state)
	}
}
func (s status) V2alpha1Status() (processingStatus.ReconciliationV2alpha1Status, error) {
	switch s.state {
	case Ready:
		return processingStatus.ReconciliationV2alpha1Status{
			ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
				State:       gatewayv2alpha1.Ready,
				Description: s.description,
			},
		}, nil
	case Error:
		return processingStatus.ReconciliationV2alpha1Status{
			ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
				State:       gatewayv2alpha1.Error,
				Description: s.description,
			},
		}, nil
	case Warning:
		return processingStatus.ReconciliationV2alpha1Status{
			ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
				State:       gatewayv2alpha1.Warning,
				Description: s.description,
			},
		}, nil
	default:
		return processingStatus.ReconciliationV2alpha1Status{}, fmt.Errorf("unsupported status: %v", s.state)
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

func (s status) Condition() *metav1.Condition {
	return s.condition
}
