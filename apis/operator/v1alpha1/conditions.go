package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Condition type constants.
const (
	ConditionTypeReady         = "Ready"
	ConditionTypeKymaGateway   = "KymaGatewayReady"
	ConditionTypeDNSEntry      = "DNSReady"
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
	ReasonDNSEntryReconcilePending      = "DNSEntryReconcilePending"
	ReasonDNSEntryReconcileFailed       = "DNSEntryReconcileFailed"
	ReasonDNSLocalDomain                = "DNSLocalDomain"
	ReasonCertificateReconcileSucceeded = "CertificateReconcileSucceeded"
	ReasonCertificateReconcilePending   = "CertificateReconcilePending"
	ReasonCertificateReconcileFailed    = "CertificateReconcileFailed"
	ReasonCertificateSelfSigned         = "CertificateSelfSigned"
)

// Reason constants for the NetworkPolicyReady condition.
const (
	ReasonNetworkPolicyReconcileSucceeded = "NetworkPolicyReconcileSucceeded"
	ReasonNetworkPolicyReconcileFailed    = "NetworkPolicyReconcileFailed"
)

// Reason constants for the DependenciesReady condition.
const (
	ReasonDependenciesReconcileSucceeded = "DependenciesReconcileSucceeded"
	ReasonDependenciesMissing            = "DependenciesMissing"
	ReasonDependenciesReconcileFailed    = "DependenciesReconcileFailed"
)

func newCondition(conditionType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
}

// Ready condition constructors.

// ProcessingCondition returns a Ready=False condition indicating reconciliation is in progress.
func ProcessingCondition() metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, ReasonReconcileProcessing, "Reconcile processing")
}

// ReadyCondition returns a Ready=True condition indicating reconciliation succeeded.
func ReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonReconcileSucceeded, "Reconciliation succeeded")
}

// ErrorCondition returns a Ready=False condition with the given reason and message.
func ErrorCondition(reason, message string) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, reason, message)
}

// DeletionBlockedExistingResourcesCondition returns a Ready=False condition indicating deletion is blocked by existing resources.
func DeletionBlockedExistingResourcesCondition(message string) metav1.Condition {
	return ErrorCondition(ReasonDeletionBlockedExistingResources, message)
}

// KymaGatewayReady condition constructors.

// KymaGatewayReadyCondition returns a KymaGatewayReady=True condition indicating the gateway reconciled successfully.
func KymaGatewayReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionTrue, ReasonKymaGatewayReconcileSucceeded, "Kyma Gateway reconciliation succeeded")
}

// KymaGatewayErrorCondition returns a KymaGatewayReady=False condition indicating a reconciliation failure.
func KymaGatewayErrorCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionFalse, ReasonKymaGatewayReconcileFailed, message)
}

// KymaGatewayDeletionBlockedCondition returns a KymaGatewayReady=False condition indicating deletion is blocked by existing resources.
func KymaGatewayDeletionBlockedCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeKymaGateway, metav1.ConditionFalse, ReasonKymaGatewayDeletionBlocked, message)
}

// Oathkeeper condition constructors. These reuse ConditionTypeReady instead of a dedicated type,
// since the legacy Oathkeeper path is slated for deprecation. Only the error reason is observable;
// the ready/disabled reasons are overwritten by the terminal ReconcileSucceeded write.

// OathkeeperReadyCondition returns a Ready=True condition indicating Oathkeeper reconciliation succeeded.
func OathkeeperReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonOathkeeperReconcileSucceeded, "Ory Oathkeeper reconciliation succeeded")
}

// OathkeeperDisabledCondition returns a Ready=True condition indicating Oathkeeper reconciliation is disabled.
func OathkeeperDisabledCondition() metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionTrue, ReasonOathkeeperReconcileDisabled, "Ory Oathkeeper reconciliation disabled")
}

// OathkeeperErrorCondition returns a Ready=False condition indicating an Oathkeeper reconciliation failure.
func OathkeeperErrorCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeReady, metav1.ConditionFalse, ReasonOathkeeperReconcileFailed, message)
}

// DNSReady condition constructors.

// DNSEntryReadyCondition returns a DNSReady=True condition indicating the Gardener DNS entry was provisioned successfully.
func DNSEntryReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionTrue, ReasonDNSEntryReconcileSucceeded, "DNSEntry reconciliation succeeded")
}

// DNSLocalDomainCondition returns a DNSReady=True condition for non-Gardener clusters using the local domain.
func DNSLocalDomainCondition() metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionTrue, ReasonDNSLocalDomain, "Using local.kyma.dev domain")
}

// DNSEntryProcessingCondition returns a DNSReady=False condition indicating the DNS entry is still being provisioned.
func DNSEntryProcessingCondition() metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionFalse, ReasonDNSEntryReconcilePending, "DNSEntry reconciliation in progress")
}

// DNSEntryErrorCondition returns a DNSReady=False condition indicating a DNS entry provisioning failure.
func DNSEntryErrorCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeDNSEntry, metav1.ConditionFalse, ReasonDNSEntryReconcileFailed, message)
}

// CertificateReady condition constructors.

// CertificateReadyCondition returns a CertificateReady=True condition indicating the Gardener certificate was issued successfully.
func CertificateReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionTrue, ReasonCertificateReconcileSucceeded, "Certificate reconciliation succeeded")
}

// CertificateSelfSignedCondition returns a CertificateReady=True condition for non-Gardener clusters using a self-signed certificate.
func CertificateSelfSignedCondition() metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionTrue, ReasonCertificateSelfSigned, "Self-signed certificate secret applied successfully")
}

// CertificateProcessingCondition returns a CertificateReady=False condition indicating the certificate is still being issued.
func CertificateProcessingCondition() metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionFalse, ReasonCertificateReconcilePending, "Certificate reconciliation in progress")
}

// CertificateErrorCondition returns a CertificateReady=False condition indicating a certificate issuance failure.
func CertificateErrorCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeCertificate, metav1.ConditionFalse, ReasonCertificateReconcileFailed, message)
}

// NetworkPolicyReady condition constructors.

// NetworkPolicyReadyCondition returns a NetworkPolicyReady=True condition indicating the NetworkPolicy reconciled successfully.
func NetworkPolicyReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeNetworkPolicy, metav1.ConditionTrue, ReasonNetworkPolicyReconcileSucceeded, "NetworkPolicy reconciliation succeeded")
}

// NetworkPolicyErrorCondition returns a NetworkPolicyReady=False condition indicating a NetworkPolicy reconciliation failure.
func NetworkPolicyErrorCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeNetworkPolicy, metav1.ConditionFalse, ReasonNetworkPolicyReconcileFailed, message)
}

// DependenciesReady condition constructors.

// DependenciesReadyCondition returns a DependenciesReady=True condition indicating all required CRDs are present.
func DependenciesReadyCondition() metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionTrue, ReasonDependenciesReconcileSucceeded, "Module dependencies available")
}

// DependenciesMissingCondition returns a DependenciesReady=False condition indicating one or more required CRDs are absent.
func DependenciesMissingCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesMissing, message)
}

// DependenciesErrorCondition returns a DependenciesReady=False condition indicating an error during dependency discovery.
func DependenciesErrorCondition(message string) metav1.Condition {
	return newCondition(ConditionTypeDependencies, metav1.ConditionFalse, ReasonDependenciesReconcileFailed, message)
}
