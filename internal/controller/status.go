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
	SetConditions(conditions ...metav1.Condition)
}

type status struct {
	err         error
	description string
	state       State
	conditions  []metav1.Condition
}

// ErrorStatus returns a Status representing a reconciliation failure with the given conditions.
func ErrorStatus(err error, description string, conditions ...metav1.Condition) Status {
	return &status{
		err:         err,
		description: description,
		state:       Error,
		conditions:  conditions,
	}
}

// WarningStatus returns a Status representing a non-fatal reconciliation issue with the given conditions.
func WarningStatus(err error, description string, conditions ...metav1.Condition) Status {
	return &status{
		err:         err,
		description: description,
		state:       Warning,
		conditions:  conditions,
	}
}

// ReadyStatus returns a Status representing successful reconciliation with the given conditions.
func ReadyStatus(conditions ...metav1.Condition) Status {
	return &status{
		description: "Successfully reconciled",
		state:       Ready,
		conditions:  conditions,
	}
}

// DeletingStatus returns a Status representing an in-progress deletion with the given conditions.
func DeletingStatus(conditions ...metav1.Condition) Status {
	return &status{
		state:      Deleting,
		conditions: conditions,
	}
}

// ProcessingStatus returns a Status representing an in-progress reconciliation with the given conditions.
func ProcessingStatus(conditions ...metav1.Condition) Status {
	return &status{
		state:      Processing,
		conditions: conditions,
	}
}

func (s *status) NestedError() error {
	return s.err
}

func (s *status) Description() string {
	return s.description
}

func (s *status) ToAPIGatewayStatus() (operatorv1alpha1.APIGatewayStatus, error) {
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

func (s *status) V2alpha1Status() (processingStatus.ReconciliationV2alpha1Status, error) {
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

func (s *status) V1beta1Status() (processingStatus.ReconciliationV1beta1Status, error) {
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

func (s *status) IsError() bool {
	return s.state == Error
}

func (s *status) IsWarning() bool {
	return s.state == Warning
}

func (s *status) IsReady() bool {
	return s.state == Ready
}

func (s *status) State() State {
	return s.state
}

// Conditions returns the conditions held by this status.
func (s *status) Conditions() []metav1.Condition {
	return s.conditions
}

// SetConditions adds conditions for types not already owned by this status — own conditions win.
func (s *status) SetConditions(conditions ...metav1.Condition) {
	own := make(map[string]struct{}, len(s.conditions))
	for _, c := range s.conditions {
		own[c.Type] = struct{}{}
	}
	for _, c := range conditions {
		if _, exists := own[c.Type]; !exists {
			meta.SetStatusCondition(&s.conditions, c)
		}
	}
}

// UpdateApiGatewayStatus writes the given Status to the APIGateway CR, stamping ObservedGeneration
// on all conditions. During Processing, existing condition ObservedGenerations are preserved to
// avoid advancing the generation counter while reconciliation is in-flight.
func UpdateApiGatewayStatus(ctx context.Context, k8sClient client.Client, apiGatewayCR *operatorv1alpha1.APIGateway, status Status) error {
	newStatus, err := status.ToAPIGatewayStatus()
	if err != nil {
		return err
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := k8sClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); getErr != nil {
			return getErr
		}

		conditions := make([]metav1.Condition, len(newStatus.Conditions))
		copy(conditions, newStatus.Conditions)

		if newStatus.State == operatorv1alpha1.Processing {
			prevByType := make(map[string]metav1.Condition, len(apiGatewayCR.Status.Conditions))
			for _, c := range apiGatewayCR.Status.Conditions {
				prevByType[c.Type] = c
			}
			for i := range conditions {
				if prev, ok := prevByType[conditions[i].Type]; ok {
					conditions[i].ObservedGeneration = prev.ObservedGeneration
				}
			}
		} else {
			for i := range conditions {
				conditions[i].ObservedGeneration = apiGatewayCR.Generation
			}
		}

		prevState := apiGatewayCR.Status.State
		prevDescription := apiGatewayCR.Status.Description

		apiGatewayCR.Status.State = newStatus.State
		apiGatewayCR.Status.Description = newStatus.Description

		var conditionsChanged bool
		for _, c := range conditions {
			if meta.SetStatusCondition(&apiGatewayCR.Status.Conditions, c) {
				conditionsChanged = true
			}
		}

		if prevState == newStatus.State && prevDescription == newStatus.Description && !conditionsChanged {
			return nil
		}
		return k8sClient.Status().Update(ctx, apiGatewayCR)
	})
}
