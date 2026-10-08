package controller

import (
	"context"
	"fmt"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
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
	V1beta1Status() (processingStatus.ReconciliationV1beta1Status, error)
	V2alpha1Status() (processingStatus.ReconciliationV2alpha1Status, error)
	IsReady() bool
	IsWarning() bool
	IsError() bool
	State() State
	Description() string
	Conditions() []metav1.Condition
	WithConditions(conditions ...metav1.Condition) Status
}

type status struct {
	err         error
	description string
	state       State
	conditions  []metav1.Condition
}

func ErrorStatus(err error, description string, conditions ...metav1.Condition) Status {
	return status{
		err:         err,
		description: description,
		state:       Error,
		conditions:  conditions,
	}
}

func WarningStatus(err error, description string, conditions ...metav1.Condition) Status {
	return status{
		err:         err,
		description: description,
		state:       Warning,
		conditions:  conditions,
	}
}

func ReadyStatus(conditions ...metav1.Condition) Status {
	return status{
		description: "Successfully reconciled",
		state:       Ready,
		conditions:  conditions,
	}
}

func DeletingStatus(conditions ...metav1.Condition) Status {
	return status{
		state:      Deleting,
		conditions: conditions,
	}
}

func ProcessingStatus(conditions ...metav1.Condition) Status {
	return status{
		state:      Processing,
		conditions: conditions,
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
	for _, c := range s.conditions {
		meta.SetStatusCondition(&newStatus.Conditions, c)
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

func (s status) V1beta1Status() (processingStatus.ReconciliationV1beta1Status, error) {
	switch s.state {
	case Ready:
		return processingStatus.ReconciliationV1beta1Status{
			ApiRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
				Code:        gatewayv1beta1.StatusOK,
				Description: s.description,
			},
		}, nil
	case Error:
		return processingStatus.ReconciliationV1beta1Status{
			ApiRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
				Code:        gatewayv1beta1.StatusError,
				Description: s.description,
			},
		}, nil
	case Warning:
		return processingStatus.ReconciliationV1beta1Status{
			ApiRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
				Code:        gatewayv1beta1.StatusWarning,
				Description: s.description,
			},
		}, nil
	default:
		return processingStatus.ReconciliationV1beta1Status{}, fmt.Errorf("unsupported status: %v", s.state)
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

func (s status) Conditions() []metav1.Condition {
	return s.conditions
}

func (s status) WithConditions(conditions ...metav1.Condition) Status {
	mergedConditions := make([]metav1.Condition, 0, len(s.conditions)+len(conditions))
	for _, condition := range conditions {
		meta.SetStatusCondition(&mergedConditions, condition)
	}
	for _, condition := range s.conditions {
		meta.SetStatusCondition(&mergedConditions, condition)
	}
	s.conditions = mergedConditions
	return s
}

func UpdateApiGatewayStatus(ctx context.Context, k8sClient client.Client, apiGatewayCR *operatorv1alpha1.APIGateway, status Status) error {
	newStatus, err := status.ToAPIGatewayStatus()
	if err != nil {
		return err
	}
	state := newStatus.State
	description := newStatus.Description
	newConditions := newStatus.Conditions
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := k8sClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); getErr != nil {
			return getErr
		}
		prevState := apiGatewayCR.Status.State
		prevDescription := apiGatewayCR.Status.Description
		prevConditions := make(map[string]metav1.Condition, len(apiGatewayCR.Status.Conditions))
		for _, c := range apiGatewayCR.Status.Conditions {
			prevConditions[c.Type] = c
		}

		apiGatewayCR.Status.State = state
		apiGatewayCR.Status.Description = description
		if state == operatorv1alpha1.Processing {
			for _, c := range newConditions {
				c.ObservedGeneration = 0
				if prev, ok := prevConditions[c.Type]; ok {
					c.ObservedGeneration = prev.ObservedGeneration
				}
				meta.SetStatusCondition(&apiGatewayCR.Status.Conditions, c)
			}
		} else {
			for _, c := range newConditions {
				c.ObservedGeneration = apiGatewayCR.Generation
				meta.SetStatusCondition(&apiGatewayCR.Status.Conditions, c)
			}
		}

		if prevState == state && prevDescription == description && conditionsUnchanged(prevConditions, apiGatewayCR.Status.Conditions) {
			return nil
		}
		return k8sClient.Status().Update(ctx, apiGatewayCR)
	})
}

func conditionsUnchanged(prev map[string]metav1.Condition, current []metav1.Condition) bool {
	if len(prev) != len(current) {
		return false
	}
	for _, c := range current {
		p, ok := prev[c.Type]
		if !ok || p.Status != c.Status || p.Reason != c.Reason || p.Message != c.Message || p.ObservedGeneration != c.ObservedGeneration {
			return false
		}
	}
	return true
}
