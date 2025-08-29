package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ReconcileProcessing              = ReasonMessage{"ReconcileProcessing", "Reconcile processing", metav1.ConditionUnknown}
	ReconcileSucceeded               = ReasonMessage{"ReconcileSucceeded", "Reconciliation succeeded", metav1.ConditionTrue}
	ReconcileFailed                  = ReasonMessage{"ReconcileFailed", "Reconciliation failed", metav1.ConditionFalse}
	OlderCRExists                    = ReasonMessage{"OlderCRExists", "API Gateway CR is not the oldest one and does not represent the module state", metav1.ConditionFalse}
	CustomResourceMisconfigured      = ReasonMessage{"CustomResourceMisconfigured", "API Gateway CR has invalid configuration", metav1.ConditionFalse}
	DependenciesMissing              = ReasonMessage{"DependenciesMissing", "Module dependencies missing", metav1.ConditionFalse}
	KymaGatewayReconcileSucceeded    = ReasonMessage{"KymaGatewayReconcileSucceeded", "Kyma Gateway reconciliation succeeded", metav1.ConditionFalse}
	KymaGatewayReconcileFailed       = ReasonMessage{"KymaGatewayReconcileFailed", "Kyma Gateway reconciliation failed", metav1.ConditionFalse}
	KymaGatewayDeletionBlocked       = ReasonMessage{"KymaGatewayDeletionBlocked", "Kyma Gateway deletion blocked because of the existing custom resources", metav1.ConditionFalse}
	OathkeeperReconcileSucceeded     = ReasonMessage{"OathkeeperReconcileSucceeded", "Ory Oathkeeper reconciliation succeeded", metav1.ConditionFalse}
	OathkeeperReconcileFailed        = ReasonMessage{"OathkeeperReconcileFailed", "Ory Oathkeeper reconciliation failed", metav1.ConditionFalse}
	OathkeeperReconcileDisabled      = ReasonMessage{"OathkeeperReconcileDisabled", "Ory Oathkeeper reconciliation disabled", metav1.ConditionFalse}
	DeletionBlockedExistingResources = ReasonMessage{"DeletionBlockedExistingResources", "API Gateway deletion blocked because of the existing custom resources", metav1.ConditionFalse}
)

// ReasonMessage is a struct that defines different states of Ready condition
type ReasonMessage struct {
	reason, message string
	status          metav1.ConditionStatus
}

// Condition returns metav1.Condition from existing ReasonMessage
func (rm ReasonMessage) Condition() *metav1.Condition {
	return &metav1.Condition{
		Type:    "Ready",
		Reason:  rm.reason,
		Message: rm.message,
		Status:  rm.status,
	}
}

// AdditionalMessage adds additional string message to already defined message field in ReasonMessage
// and returns a new ReasonMessage based on parent
func (rm ReasonMessage) AdditionalMessage(message string) ReasonMessage {
	rm.message = rm.message + message
	return rm
}
