package v2alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Condition type constants.
const (
	ConditionTypeReady = "Ready"
)

// Condition reason constants.
const (
	ReasonReconcileSucceeded                   = "ReconcileSucceeded"
	ReasonVirtualServiceReconcileFailed        = "VirtualServiceReconcileFailed"
	ReasonRequestAuthenticationReconcileFailed = "RequestAuthenticationReconcileFailed"
	ReasonAuthorizationPolicyReconcileFailed   = "AuthorizationPolicyReconcileFailed"
	ReasonReconcileFailed                      = "ReconcileFailed"
	ReasonValidationFailed                     = "ValidationFailed"
	ReasonGatewayNotFound                      = "GatewayNotFound"
	ReasonDependenciesMissing                  = "DependenciesMissing"
)

func newCondition(status metav1.ConditionStatus, reason, message string, generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
	}
}

func ReadyCondition(generation int64) metav1.Condition {
	return newCondition(metav1.ConditionTrue, ReasonReconcileSucceeded, "Reconciled successfully", generation)
}

func ErrorConditionWithReason(generation int64, reason, message string) metav1.Condition {
	return newCondition(metav1.ConditionFalse, reason, message, generation)
}

func ErrorCondition(generation int64, message string) metav1.Condition {
	return ErrorConditionWithReason(generation, ReasonReconcileFailed, message)
}
