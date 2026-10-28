package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Condition type constants.
const (
	ConditionTypeReady        = "Ready"
	ConditionTypeDependencies = "DependenciesReady"
)

// Reason constants for the Ready condition.
const (
	ReasonReconcileSucceeded         = "ReconcileSucceeded"
	ReasonReconcileFailed            = "ReconcileFailed"
	ReasonValidationFailed           = "ValidationFailed"
	ReasonAPIGatewayNotReady         = "APIGatewayNotReady"
	ReasonEnvoyFilterReconcileFailed = "EnvoyFilterReconcileFailed"
)

// Reason constants for the DependenciesReady condition.
const (
	ReasonDependenciesAvailable       = "DependenciesReconcileSucceeded"
	ReasonDependenciesMissing         = "DependenciesMissing"
	ReasonDependenciesReconcileFailed = "DependenciesReconcileFailed"
)

func newCondition(conditionType string, status metav1.ConditionStatus, reason, message string, generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: generation,
	}
}

// Ready condition constructors.

func ReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonReconcileSucceeded, "Reconciled successfully", generation)
}

func ErrorCondition(generation int64, reason, message string) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, reason, message, generation)
}

// DependenciesReady condition constructors.

func DependenciesReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionTrue, ReasonDependenciesAvailable, "Module dependencies available", generation)
}

func DependenciesMissingCondition(generation int64, message string) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesMissing, message, generation)
}

func DependenciesErrorCondition(generation int64, message string) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesReconcileFailed, message, generation)
}
