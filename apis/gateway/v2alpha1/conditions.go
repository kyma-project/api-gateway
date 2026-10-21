package v2alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Condition type constants.
const (
	ConditionTypeReady                  = "Ready"
	ConditionTypeVirtualService         = "VirtualServiceConfigured"
	ConditionTypeRequestAuthentication  = "RequestAuthenticationConfigured"
	ConditionTypeAuthorizationPolicy    = "AuthorizationPolicyConfigured"
	ConditionTypeSubresourcesConfigured = "SubresourcesConfigured"
	ConditionTypeDependencies           = "DependenciesReady"
)

// Condition reason constants.
const (
	ReasonReconcileSucceeded               = "ReconcileSucceeded"
	ReasonReconcileProcessing              = "NotYetReady"
	ReasonReconcileFailed                  = "ReconcileFailed"
	ReasonDeletionBlockedExistingResources = "DeletionBlockedExistingResources"
	ReasonValidationFailed                 = "ValidationFailed"
	ReasonGatewayNotFound                  = "GatewayNotFound"
	ReasonDependenciesMissing              = "DependenciesMissing"
	ReasonDependenciesAvailable            = "DependenciesReconcileSucceeded"
	ReasonDependenciesError                = "DependenciesError"

	ReasonVirtualServiceReconcileSucceeded         = "VirtualServiceReconcileSucceeded"
	ReasonVirtualServiceReconcileProcessing        = "VirtualServiceNotYetReady"
	ReasonVirtualServiceReconcileFailed            = "VirtualServiceReconcileFailed"
	ReasonRequestAuthenticationReconcileSucceeded  = "RequestAuthenticationReconcileSucceeded"
	ReasonRequestAuthenticationReconcileProcessing = "RequestAuthenticationNotYetReady"
	ReasonRequestAuthenticationReconcileFailed     = "RequestAuthenticationReconcileFailed"
	ReasonAuthorizationPolicyReconcileSucceeded    = "AuthorizationPolicyReconcileSucceeded"
	ReasonAuthorizationPolicyReconcileProcessing   = "AuthorizationPolicyNotYetReady"
	ReasonAuthorizationPolicyReconcileFailed       = "AuthorizationPolicyReconcileFailed"
	ReasonSubresourcesReconcileSucceeded           = "SubresourcesReconcileSucceeded"
	ReasonSubresourcesReconcileProcessing          = "SubresourcesNotYetReady"
	ReasonSubresourcesReconcileFailed              = "SubresourcesReconcileFailed"
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
	return newCondition(conditionType, metav1.ConditionUnknown, reason, message, generation)
}

func errorCondition(conditionType, reason, message string, generation int64) metav1.Condition {
	return newCondition(conditionType, metav1.ConditionFalse, reason, message, generation)
}

// Ready condition constructors.

func ProcessingCondition() metav1.Condition {
	return ProcessingConditionWithGeneration(0)
}

func ProcessingConditionWithGeneration(generation int64) metav1.Condition {
	return processingCondition(ConditionTypeReady, ReasonReconcileProcessing, "Reconcile processing", generation)
}

func ReadyCondition() metav1.Condition {
	return ReadyConditionWithGeneration(0)
}

func ReadyConditionWithGeneration(generation int64) metav1.Condition {
	return configuredCondition(ConditionTypeReady, ReasonReconcileSucceeded, "Reconciled successfully", generation)
}

func ErrorCondition(reason, message string) metav1.Condition {
	return ErrorConditionWithGeneration(0, reason, message)
}

func ErrorConditionWithGeneration(generation int64, reason, message string) metav1.Condition {
	return errorCondition(ConditionTypeReady, reason, message, generation)
}

func WarningCondition(reason, message string) metav1.Condition {
	return errorCondition(ConditionTypeReady, reason, message, 0)
}

func DeletionBlockedExistingResourcesCondition(message string) metav1.Condition {
	return WarningCondition(ReasonDeletionBlockedExistingResources, message)
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

// SubresourcesConfigured condition constructors.

func SubresourcesConfiguredCondition(generation int64) metav1.Condition {
	return configuredCondition(ConditionTypeSubresourcesConfigured, ReasonSubresourcesReconcileSucceeded, "Subresources reconciliation succeeded", generation)
}

func SubresourcesProcessingCondition(generation int64) metav1.Condition {
	return processingCondition(ConditionTypeSubresourcesConfigured, ReasonSubresourcesReconcileProcessing, "Subresources reconciliation in progress", generation)
}

func SubresourcesErrorCondition(generation int64, message string) metav1.Condition {
	return errorCondition(ConditionTypeSubresourcesConfigured, ReasonSubresourcesReconcileFailed, message, generation)
}

// DependenciesReady condition constructors.

func DependenciesReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionTrue, ReasonDependenciesAvailable, "Module dependencies available", 0)
}

func DependenciesMissingCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesMissing, message, 0)
}

func DependenciesErrorCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesError, message, 0)
}

// ProcessingConditions returns the full set of per-resource conditions set to
// Unknown, signalling that reconciliation is in progress and no sub-resource has
// been evaluated yet this pass.
func ProcessingConditions(generation int64) []metav1.Condition {
	return []metav1.Condition{
		ProcessingConditionWithGeneration(generation),
		VirtualServiceProcessingCondition(generation),
		RequestAuthenticationProcessingCondition(generation),
		AuthorizationPolicyProcessingCondition(generation),
		SubresourcesProcessingCondition(generation),
	}
}
