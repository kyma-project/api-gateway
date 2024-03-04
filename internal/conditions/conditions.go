package conditions

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ReconcileProcessing              = ReasonMessage{"ReconcileProcessing", "Reconcile processing", metav1.ConditionUnknown}
	ReconcileSucceeded               = ReasonMessage{"ReconcileSucceeded", "Reconciliation succeeded", metav1.ConditionTrue}
	ReconcileFailed                  = ReasonMessage{"ReconcileFailedReason", "Reconciliation failed", metav1.ConditionFalse}
	OlderCRExists                    = ReasonMessage{"OlderCRExistsReason", "Reconciliation failed", metav1.ConditionFalse}
	CustomResourceMisconfigured      = ReasonMessage{"CustomResourceMisconfiguredReason", "API Gateway CR has invalid configuration", metav1.ConditionFalse}
	DependenciesMissing              = ReasonMessage{"DependenciesMissingReason", "Module dependencies missing", metav1.ConditionFalse}
	KymaGatewayReconcileSucceeded    = ReasonMessage{"KymaGatewayReconcileSucceededReason", "Kyma Gateway reconciliation succeeded", metav1.ConditionFalse}
	KymaGatewayReconcileFailed       = ReasonMessage{"KymaGatewayReconcileFailedReason", "Kyma Gateway reconciliation failed", metav1.ConditionFalse}
	KymaGatewayDeletionBlocked       = ReasonMessage{"KymaGatewayDeletionBlockedReason", "Kyma Gateway deletion blocked because of the existing custom resources", metav1.ConditionFalse}
	OathkeeperReconcileSucceeded     = ReasonMessage{"OathkeeperReconcileSucceeded", "Ory Oathkeeper reconciliation succeeded", metav1.ConditionFalse}
	OathkeeperReconcileFailed        = ReasonMessage{"OathkeeperReconcileFailed", "Ory Oathkeeper reconciliation failed", metav1.ConditionFalse}
	DeletionBlockedExistingResources = ReasonMessage{"DeletionBlockedExistingResources", "API Gateway deletion blocked because of the existing custom resources", metav1.ConditionFalse}
)

type ReasonMessage struct {
	reason, message string
	status          metav1.ConditionStatus
}

func (rm *ReasonMessage) Condition() *metav1.Condition {
	return &metav1.Condition{
		Type:    "Ready",
		Reason:  rm.reason,
		Message: rm.message,
		Status:  rm.status,
	}
}

func (rm *ReasonMessage) AdditionalMessage(message string) *ReasonMessage {
	rm.message = fmt.Sprintf(rm.message, message)
	return rm
}
