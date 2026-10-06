package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Condition type constants.
const (
	ConditionTypeReady       = "Ready"
	ConditionTypeKymaGateway = "KymaGatewayReady"
	//BEZ TEGO ConditionTypeOathkeeper    = "OathkeeperReady"
	ConditionTypeNetworkPolicy = "NetworkPolicyReady"
	ConditionTypeDependencies  = "DependenciesReady"
)

// Reason constants for the Ready condition.
const (
	ReasonReconcileProcessing              = "ReconcileProcessing"
	ReasonReconcileSucceeded               = "ReconcileSucceeded"
	ReasonReconcileFailed                  = "ReconcileFailed"
	ReasonOlderCRExists                    = "OlderCRExists"
	ReasonCustomResourceMisconfigured      = "CustomResourceMisconfigured"
	ReasonDeletionBlockedExistingResources = "DeletionBlockedExistingResources"
	ReasonOathkeeperReconcileSucceeded     = "OathkeeperReconcileSucceeded"
	ReasonOathkeeperReconcileFailed        = "OathkeeperReconcileFailed"
	ReasonOathkeeperReconcileDisabled      = "OathkeeperReconcileDisabled"
)

// Reason constants for the GatewayReady condition.
const (
	ReasonKymaGatewayReconcileSucceeded = "KymaGatewayReconcileSucceeded"
	ReasonKymaGatewayReconcileFailed    = "KymaGatewayReconcileFailed"
	ReasonKymaGatewayDeletionBlocked    = "KymaGatewayDeletionBlocked"
)

// Reason constants for the NetworkPolicyReady condition.
const (
	ReasonNetworkPolicyReconcileSucceeded = "NetworkPolicyReconcileSucceeded"
	ReasonNetworkPolicyReconcileFailed    = "NetworkPolicyReconcileFailed"
)

// Reason constants for the DependenciesReady condition.
const (
	ReasonDependenciesAvailable = "Available"
	ReasonDependenciesMissing   = "DependenciesMissing"
	ReasonDependenciesError     = "DependenciesError"
)

// Ready condition constructors.

func ProcessingCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionUnknown,
		Reason:  ReasonReconcileProcessing,
		Message: "Reconcile processing",
	}
}

func ReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonReconcileSucceeded,
		Message: "Reconciliation succeeded",
	}
}

func ErrorCondition(reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	}
}

func WarningCondition(reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	}
}

// GatewayReady condition constructors.

func KymaGatewayReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeKymaGateway,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonKymaGatewayReconcileSucceeded,
		Message: "Kyma Gateway reconciliation succeeded",
	}
}

func KymaGatewayProcessingCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeKymaGateway,
		Status:  metav1.ConditionUnknown,
		Reason:  ReasonKymaGatewayReconcileFailed,
		Message: "Kyma Gateway reconciliation in progress",
	}
}

func KymaGatewayErrorCondition(message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeKymaGateway,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonKymaGatewayReconcileFailed,
		Message: message,
	}
}

func KymaGatewayDeletionBlockedCondition(message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeKymaGateway,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonKymaGatewayDeletionBlocked,
		Message: message,
	}
}

// OathkeeperReady condition constructors.

func OathkeeperReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonOathkeeperReconcileSucceeded,
		Message: "Ory Oathkeeper reconciliation succeeded",
	}
}

func OathkeeperDisabledCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonOathkeeperReconcileDisabled,
		Message: "Ory Oathkeeper reconciliation disabled",
	}
}

func OathkeeperErrorCondition(message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonOathkeeperReconcileFailed,
		Message: message,
	}
}

// NetworkPolicyReady condition constructors.

func NetworkPolicyReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeNetworkPolicy,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonNetworkPolicyReconcileSucceeded,
		Message: "NetworkPolicy reconciliation succeeded",
	}
}

func NetworkPolicyErrorCondition(message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeNetworkPolicy,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonNetworkPolicyReconcileFailed,
		Message: message,
	}
}

// DependenciesReady condition constructors.

func DependenciesReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDependencies,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonDependenciesAvailable,
		Message: "Module dependencies available",
	}
}

func DependenciesMissingCondition(message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDependencies,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonDependenciesMissing,
		Message: message,
	}
}

func DependenciesErrorCondition(message string) metav1.Condition {
	return metav1.Condition{
		Type:    ConditionTypeDependencies,
		Status:  metav1.ConditionFalse,
		Reason:  ReasonDependenciesError,
		Message: message,
	}
}
