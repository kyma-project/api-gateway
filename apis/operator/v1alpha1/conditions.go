package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Condition type constants.
const (
	ConditionTypeReady         = "Ready"
	ConditionTypeKymaGateway   = "KymaGatewayReady"
	ConditionTypeDNSEntry      = "DNSEntryReady"
	ConditionTypeCertificate   = "CertificateReady"
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
	ReasonKymaGatewayReconcileSucceeded  = "KymaGatewayReconcileSucceeded"
	ReasonKymaGatewayReconcileDisabled   = "KymaGatewayReconcileDisabled"
	ReasonKymaGatewayReconcileProcessing = "KymaGatewayReconcileProcessing"
	ReasonKymaGatewayReconcileFailed     = "KymaGatewayReconcileFailed"
	ReasonKymaGatewayDeletionBlocked     = "KymaGatewayDeletionBlocked"
)

// Reason constants for DNSEntry and Certificate conditions.
const (
	ReasonDNSEntryReconcileSucceeded    = "DNSEntryReconcileSucceeded"
	ReasonDNSEntryReconcilePending      = "DNSEntryReconcilePending"
	ReasonDNSEntryReconcileFailed       = "DNSEntryReconcileFailed"
	ReasonCertificateReconcileSucceeded = "CertificateReconcileSucceeded"
	ReasonCertificateReconcilePending   = "CertificateReconcilePending"
	ReasonCertificateReconcileFailed    = "CertificateReconcileFailed"
)

// Reason constants for the NetworkPolicyReady condition.
const (
	ReasonNetworkPolicyReconcileSucceeded = "NetworkPolicyReconcileSucceeded"
	ReasonNetworkPolicyReconcileDisabled  = "NetworkPolicyReconcileDisabled"
	ReasonNetworkPolicyReconcileFailed    = "NetworkPolicyReconcileFailed"
)

// Reason constants for the DependenciesReady condition.
const (
	ReasonDependenciesAvailable = "DependenciesReconcileSucceeded"
	ReasonDependenciesMissing   = "DependenciesMissing"
	ReasonDependenciesError     = "DependenciesReconcileFailed"
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

func ProcessingCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionUnknown, ReasonReconcileProcessing, "Reconcile processing", generation)
}

func ReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonReconcileSucceeded, "Reconciliation succeeded", generation)
}

func ErrorCondition(reason, message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, reason, message, generation)
}

func WarningCondition(reason, message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, reason, message, generation)
}

func DeletionBlockedExistingResourcesCondition(message string, generation int64) metav1.Condition {
	return WarningCondition(ReasonDeletionBlockedExistingResources, message, generation)
}

// KymaGatewayReady condition constructors.

func KymaGatewayReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionTrue, ReasonKymaGatewayReconcileSucceeded, "Kyma Gateway reconciliation succeeded", generation)
}

func KymaGatewayDisabledCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionFalse, ReasonKymaGatewayReconcileDisabled, "Kyma Gateway reconciliation skipped because Kyma Gateway is disabled", generation)
}

func KymaGatewayProcessingCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionUnknown, ReasonKymaGatewayReconcileProcessing, "Kyma Gateway reconciliation in progress", generation)
}

func KymaGatewayErrorCondition(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionFalse, ReasonKymaGatewayReconcileFailed, message, generation)
}

func KymaGatewayDeletionBlockedCondition(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionFalse, ReasonKymaGatewayDeletionBlocked, message, generation)
}

// OathkeeperReady condition constructors.

func OathkeeperReconcileSucceeded(generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonOathkeeperReconcileSucceeded, "Ory Oathkeeper reconciliation succeeded", generation)
}

func OathkeeperDisabledCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonOathkeeperReconcileDisabled, "Ory Oathkeeper reconciliation disabled", generation)
}

func OathkeeperReconcileFailed(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, ReasonOathkeeperReconcileFailed, message, generation)
}

// DNSEntryReady condition constructors.

func DNSEntryReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionTrue, ReasonDNSEntryReconcileSucceeded, "DNSEntry reconciliation succeeded", generation)
}

func DNSEntryProcessingCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionFalse, ReasonDNSEntryReconcilePending, "DNSEntry reconciliation in progress", generation)
}

func DNSEntryErrorCondition(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionFalse, ReasonDNSEntryReconcileFailed, message, generation)
}

// CertificateReady condition constructors.

func CertificateReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionTrue, ReasonCertificateReconcileSucceeded, "Certificate reconciliation succeeded", generation)
}

func CertificateProcessingCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionFalse, ReasonCertificateReconcilePending, "Certificate reconciliation in progress", generation)
}

func CertificateErrorCondition(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionFalse, ReasonCertificateReconcileFailed, message, generation)
}

// NetworkPolicyReady condition constructors.

func NetworkPolicyReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeNetworkPolicy, metav1.ConditionTrue, ReasonNetworkPolicyReconcileSucceeded, "NetworkPolicy reconciliation succeeded", generation)
}

func NetworkPolicyDisabledCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeNetworkPolicy, metav1.ConditionFalse, ReasonNetworkPolicyReconcileDisabled, "NetworkPolicy reconciliation skipped because NetworkPolicies are disabled", generation)
}

func NetworkPolicyErrorCondition(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeNetworkPolicy, metav1.ConditionFalse, ReasonNetworkPolicyReconcileFailed, message, generation)
}

// DependenciesReady condition constructors.

func DependenciesReadyCondition(generation int64) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionTrue, ReasonDependenciesAvailable, "Module dependencies available", generation)
}

func DependenciesMissingCondition(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesMissing, message, generation)
}

func DependenciesErrorCondition(message string, generation int64) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesError, message, generation)
}
