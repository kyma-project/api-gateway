/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"crypto/sha256"
	"encoding/hex"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// maxBaseNameLength ensures subresources stay under 63 char Kubernetes limit
	// 63 (k8s limit) - 5 (longest suffix "-xfcc") - 1 (safety margin) = 57
	// Using 45 for extra headroom
	maxBaseNameLength = 45
)

// ExternalGatewaySpec defines the desired state of ExternalGateway
type ExternalGatewaySpec struct {
	// ExternalDomain is the customer-facing domain (e.g., api.customer.com or *.api.customer.com)
	// Supports Istio Gateway host format with optional wildcard prefix (e.g., *.example.com)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Pattern=`^(\*\.)?([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`
	ExternalDomain string `json:"externalDomain"`

	// InternalDomain configuration for Kyma-internal access
	// +kubebuilder:validation:Required
	InternalDomain InternalDomainConfig `json:"internalDomain"`

	// Region is a region identifier (e.g., "eu10", "us10")
	// This must match a btp_region defined in the external-gateway-regions ConfigMap
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// RegionsConfigMap is the name of the ConfigMap containing region metadata.
	// ConfigMap must be in the same namespace as the ExternalGateway.
	// If key is not specified in ConfigMap, auto-detects single key or looks for "regions.yaml".
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	RegionsConfigMap string `json:"regionsConfigMap"`

	// CASecretRef references the Secret containing the CA certificate
	// This CA is used to validate client certificates during mTLS handshake
	// If namespace is not specified, defaults to the ExternalGateway's namespace
	// The Secret key is not specified in Secret, auto-detects single key or looks for "ca.crt".
	// +kubebuilder:validation:Required
	CASecretRef *corev1.SecretReference `json:"caSecretRef"`
}

// InternalDomainConfig defines the Kyma-internal domain configuration
type InternalDomainConfig struct {
	// KymaSubdomain is the subdomain prefix (e.g., "external-myapp")
	// The full internal domain will be: {kymaSubdomain}.{KYMA_DOMAIN}
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	//+kubebuilder:default:="external"
	KymaSubdomain string `json:"kymaSubdomain"`
}

// State defines the reconciliation state of the ExternalGateway
type State string

const (
	// Processing the ExternalGateway is being created or updated
	Processing State = "Processing"
	// Ready the ExternalGateway's reconciliation is finished
	Ready State = "Ready"
	// Error an error occurred during reconciliation
	Error State = "Error"
)

// ExternalGatewayStatus defines the observed state of ExternalGateway
type ExternalGatewayStatus struct {
	// Represents the last time the ExternalGateway status was processed
	// +optional
	LastProcessedTime metav1.Time `json:"lastProcessedTime,omitempty"`

	// Defines the reconciliation state of the ExternalGateway
	// +kubebuilder:validation:Enum=Processing;Ready;Error
	// +optional
	State State `json:"state,omitempty"`

	// Contains the description of the ExternalGateway's status
	// +optional
	Description string `json:"description,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories=all;kyma-project;externalgateway
// +kubebuilder:printcolumn:name="External Domain",type=string,JSONPath=`.spec.externalDomain`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ExternalGateway is the Schema for the externalgateways API
type ExternalGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExternalGatewaySpec   `json:"spec,omitempty"`
	Status ExternalGatewayStatus `json:"status,omitempty"`
}

// BaseName returns a truncated name with hash suffix if the original name exceeds maxBaseNameLength.
// This ensures all derived resource names stay under Kubernetes' 63 character limit.
func (e *ExternalGateway) BaseName() string {
	if len(e.Name) <= maxBaseNameLength {
		return e.Name
	}
	// Truncate and add hash suffix for uniqueness
	// Format: {first-37-chars}-{7-char-hash} = 45 chars
	hash := sha256.Sum256([]byte(e.Name))
	hashSuffix := hex.EncodeToString(hash[:])[:7]
	return e.Name[:maxBaseNameLength-8] + "-" + hashSuffix
}

// GatewayName returns the name for the Istio Gateway resource
func (e *ExternalGateway) GatewayName() string {
	return e.BaseName() + "-gw"
}

// CertificateName returns the name for the Certificate resource
func (e *ExternalGateway) CertificateName() string {
	return e.BaseName() + "-cert"
}

// TLSSecretName returns the name for the TLS Secret
func (e *ExternalGateway) TLSSecretName() string {
	return e.BaseName() + "-tls"
}

// CASecretName returns the name for the CA Secret
func (e *ExternalGateway) CASecretName() string {
	return e.BaseName() + "-tls-cacert"
}

// DNSEntryName returns the name for the DNSEntry resource
func (e *ExternalGateway) DNSEntryName() string {
	return e.BaseName() + "-dns"
}

// XFCCFilterName returns the name for the XFCC sanitization EnvoyFilter
func (e *ExternalGateway) XFCCFilterName() string {
	return e.BaseName() + "-xfcc"
}

// CertValidationFilterName returns the name for the cert validation EnvoyFilter
func (e *ExternalGateway) CertValidationFilterName() string {
	return e.BaseName() + "-cv"
}

// +kubebuilder:object:root=true

// ExternalGatewayList contains a list of ExternalGateway
type ExternalGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExternalGateway{}, &ExternalGatewayList{})
}
