package v2alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Condition type constants.
const (
	ConditionTypeReady                 = "Ready"
	ConditionTypeVirtualService        = "VirtualServiceConfigured"
	ConditionTypeRequestAuthentication = "RequestAuthenticationConfigured"
	ConditionTypeAuthorizationPolicy   = "AuthorizationPolicyConfigured"
	ConditionTypeDependencies          = "DependenciesReady"
)

// Reason constants for the Ready condition.
const (
	ReasonReconcileSucceeded  = "ReconcileSucceeded"
	ReasonReconcileProcessing = "ReconcileProcessing"
	ReasonReconcileFailed     = "ReconcileFailed"
	ReasonValidationFailed    = "ValidationFailed"
	ReasonGatewayNotFound     = "GatewayNotFound"
)

// Reason constants for the VirtualServiceConfigured condition.
const (
	ReasonVirtualServiceReconcileSucceeded  = "VirtualServiceReconcileSucceeded"
	ReasonVirtualServiceReconcileProcessing = "VirtualServiceReconcileProcessing"
	ReasonVirtualServiceReconcileFailed     = "VirtualServiceReconcileFailed"
)

// Reason constants for the RequestAuthenticationConfigured condition.
const (
	ReasonRequestAuthenticationReconcileSucceeded  = "RequestAuthenticationReconcileSucceeded"
	ReasonRequestAuthenticationReconcileProcessing = "RequestAuthenticationReconcileProcessing"
	ReasonRequestAuthenticationReconcileFailed     = "RequestAuthenticationReconcileFailed"
)

// Reason constants for the AuthorizationPolicyConfigured condition.
const (
	ReasonAuthorizationPolicyReconcileSucceeded  = "AuthorizationPolicyReconcileSucceeded"
	ReasonAuthorizationPolicyReconcileProcessing = "AuthorizationPolicyReconcileProcessing"
	ReasonAuthorizationPolicyReconcileFailed     = "AuthorizationPolicyReconcileFailed"
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

func configuredCondition(conditionType, reason, message string, generation int64) metav1.Condition {
	return newCondition(conditionType, metav1.ConditionTrue, reason, message, generation)
}

func processingCondition(conditionType, reason, message string, generation int64) metav1.Condition {
	return newCondition(conditionType, metav1.ConditionFalse, reason, message, generation)
}

func errorCondition(conditionType, reason, message string, generation int64) metav1.Condition {
	return newCondition(conditionType, metav1.ConditionFalse, reason, message, generation)
}

// Ready condition constructors.

func ReadyCondition(generation int64) metav1.Condition {
	return configuredCondition(ConditionTypeReady, ReasonReconcileSucceeded, "Reconciled successfully", generation)
}

func ProcessingCondition(generation int64) metav1.Condition {
	return processingCondition(ConditionTypeReady, ReasonReconcileProcessing, "Reconcile processing", generation)
}

func ErrorCondition(generation int64, reason, message string) metav1.Condition {
	return errorCondition(ConditionTypeReady, reason, message, generation)
}

func GatewayNotFoundCondition(generation int64, message string) metav1.Condition {
	return errorCondition(ConditionTypeReady, ReasonGatewayNotFound, message, generation)
}

// VirtualServiceConfigured condition constructors.

func VirtualServiceConfiguredCondition(generation int64) metav1.Condition {
	return configuredCondition(ConditionTypeVirtualService, ReasonVirtualServiceReconcileSucceeded, "VirtualService reconciliation succeeded", generation)
}

func VirtualServiceProcessingCondition(generation int64) metav1.Condition {
	return processingCondition(ConditionTypeVirtualService, ReasonVirtualServiceReconcileProcessing, "VirtualService reconciliation in progress", generation)
}

func VirtualServiceErrorCondition(generation int64, message string) metav1.Condition {
	return errorCondition(ConditionTypeVirtualService, ReasonVirtualServiceReconcileFailed, message, generation)
}

// RequestAuthenticationConfigured condition constructors.

func RequestAuthenticationConfiguredCondition(generation int64) metav1.Condition {
	return configuredCondition(ConditionTypeRequestAuthentication, ReasonRequestAuthenticationReconcileSucceeded, "RequestAuthentication reconciliation succeeded", generation)
}

func RequestAuthenticationProcessingCondition(generation int64) metav1.Condition {
	return processingCondition(ConditionTypeRequestAuthentication, ReasonRequestAuthenticationReconcileProcessing, "RequestAuthentication reconciliation in progress", generation)
}

func RequestAuthenticationErrorCondition(generation int64, message string) metav1.Condition {
	return errorCondition(ConditionTypeRequestAuthentication, ReasonRequestAuthenticationReconcileFailed, message, generation)
}

// AuthorizationPolicyConfigured condition constructors.

func AuthorizationPolicyConfiguredCondition(generation int64) metav1.Condition {
	return configuredCondition(ConditionTypeAuthorizationPolicy, ReasonAuthorizationPolicyReconcileSucceeded, "AuthorizationPolicy reconciliation succeeded", generation)
}

func AuthorizationPolicyProcessingCondition(generation int64) metav1.Condition {
	return processingCondition(ConditionTypeAuthorizationPolicy, ReasonAuthorizationPolicyReconcileProcessing, "AuthorizationPolicy reconciliation in progress", generation)
}

func AuthorizationPolicyErrorCondition(generation int64, message string) metav1.Condition {
	return errorCondition(ConditionTypeAuthorizationPolicy, ReasonAuthorizationPolicyReconcileFailed, message, generation)
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

// ProcessingConditions returns the full set of per-resource conditions set to
// False/ReconcileProcessing, signalling that reconciliation is in progress and
// no sub-resource has been evaluated yet this pass.
func ProcessingConditions(generation int64) []metav1.Condition {
	return []metav1.Condition{
		ProcessingCondition(generation),
		VirtualServiceProcessingCondition(generation),
		RequestAuthenticationProcessingCondition(generation),
		AuthorizationPolicyProcessingCondition(generation),
	}
}
