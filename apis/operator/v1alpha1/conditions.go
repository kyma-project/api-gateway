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
	ReasonKymaGatewayReconcileSucceeded = "KymaGatewayReconcileSucceeded"
	ReasonKymaGatewayReconcileFailed    = "KymaGatewayReconcileFailed"
	ReasonKymaGatewayDeletionBlocked    = "KymaGatewayDeletionBlocked"
)

// Reason constants for DNSEntry and Certificate conditions.
const (
	ReasonDNSEntryReconcileSucceeded    = "DNSEntryReconcileSucceeded"
	ReasonDNSEntryReconcileFailed       = "DNSEntryReconcileFailed"
	ReasonCertificateReconcileSucceeded = "CertificateReconcileSucceeded"
	ReasonCertificateReconcilePending   = "CertificateReconcilePending"
	ReasonCertificateReconcileFailed    = "CertificateReconcileFailed"
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

func newCondition(conditionType string, status metav1.ConditionStatus, reason, message string) *metav1.Condition {
	return &metav1.Condition{
		Type:    conditionType,
		Reason:  reason,
		Message: message,
		Status:  status,
	}
}

// Ready condition constructors.

func ProcessingCondition() *metav1.Condition {
	condition := newCondition(ConditionTypeReady, metav1.ConditionUnknown, ReasonReconcileProcessing, "Reconcile processing")
	return condition
}

func ReadyCondition() *metav1.Condition {
	condition := newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonReconcileSucceeded, "Reconciliation succeeded")
	return condition
}

func ErrorCondition(reason, message string) *metav1.Condition {
	condition := newCondition(ConditionTypeReady, metav1.ConditionFalse, reason, message)
	return condition
}

func WarningCondition(reason, message string) *metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, reason, message)
}

func DeletionBlockedExistingResourcesCondition(message string) *metav1.Condition {
	return WarningCondition(ReasonDeletionBlockedExistingResources, message)
}

// GatewayReady condition constructors.

func KymaGatewayReadyCondition() *metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionTrue, ReasonKymaGatewayReconcileSucceeded, "Kyma Gateway reconciliation succeeded")
}

func KymaGatewayProcessingCondition() *metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionUnknown, ReasonKymaGatewayReconcileFailed, "Kyma Gateway reconciliation in progress")
}

func KymaGatewayErrorCondition(message string) *metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionFalse, ReasonKymaGatewayReconcileFailed, message)
}

func KymaGatewayDeletionBlockedCondition(message string) *metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionFalse, ReasonKymaGatewayDeletionBlocked, message)
}

// OathkeeperReady condition constructors.

func OathkeeperReconcileSucceeded() *metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, ReasonOathkeeperReconcileSucceeded, "Ory Oathkeeper reconciliation succeeded")
}

func OathkeeperDisabledCondition() *metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, ReasonOathkeeperReconcileDisabled, "Ory Oathkeeper reconciliation disabled")
}

func OathkeeperReconcileFailed(message string) *metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, ReasonOathkeeperReconcileFailed, message)
}

// DNSEntryReady condition constructors.

func DNSEntryReadyCondition() *metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionTrue, ReasonDNSEntryReconcileSucceeded, "DNSEntry reconciliation succeeded")
}

func DNSEntryErrorCondition(message string) *metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionFalse, ReasonDNSEntryReconcileFailed, message)
}

// CertificateReady condition constructors.

func CertificateReadyCondition() *metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionTrue, ReasonCertificateReconcileSucceeded, "Certificate reconciliation succeeded")
}

func CertificateProcessingCondition() *metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionUnknown, ReasonCertificateReconcilePending, "Certificate reconciliation in progress")
}

func CertificateErrorCondition(message string) *metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionFalse, ReasonCertificateReconcileFailed, message)
}

// NetworkPolicyReady condition constructors.

func NetworkPolicyReadyCondition() *metav1.Condition {
	return newCondition(ConditionTypeNetworkPolicy, metav1.ConditionTrue, ReasonNetworkPolicyReconcileSucceeded, "NetworkPolicy reconciliation succeeded")
}

func NetworkPolicyErrorCondition(message string) *metav1.Condition {
	return newCondition(ConditionTypeNetworkPolicy, metav1.ConditionFalse, ReasonNetworkPolicyReconcileFailed, message)
}

// DependenciesReady condition constructors.

func DependenciesReadyCondition() *metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionTrue, ReasonDependenciesAvailable, "Module dependencies available")
}

func DependenciesMissingCondition(message string) *metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesMissing, message)
}

func DependenciesErrorCondition(message string) *metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesError, message)
}
