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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExternalGatewaySpec defines the desired state of ExternalGateway
type ExternalGatewaySpec struct {
	// ExternalDomain is the customer-facing domain (e.g., api.customer.com)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Pattern=`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`
	ExternalDomain string `json:"externalDomain"`

	// InternalDomain configuration for Kyma-internal access
	// +kubebuilder:validation:Required
	InternalDomain InternalDomainConfig `json:"internalDomain"`

	// Regions is a list of UGW region identifiers (e.g., "aws/eu-central-1")
	// These must match regions defined in the external-gateway-regions ConfigMap
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Regions []string `json:"regions"`

	// Gateway is the name of the Istio Gateway to be created in the application namespace
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Gateway string `json:"gateway"`

	// CASecretRef references the Secret containing the CA certificate
	// This CA is used to validate client certificates during mTLS handshake
	// If namespace is not specified, defaults to the ExternalGateway's namespace
	// The Secret must contain a 'cacert' key (Istio convention)
	// +kubebuilder:validation:Required
	CASecretRef *corev1.SecretReference `json:"caSecretRef"`
}

// InternalDomainConfig defines the Kyma-internal domain configuration
type InternalDomainConfig struct {
	// KymaSubdomain is the subdomain prefix (e.g., "external-myapp")
	// The full internal domain will be: {kymaSubdomain}.{KYMA_DOMAIN}
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	KymaSubdomain string `json:"kymaSubdomain,default=external"`
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
// +kubebuilder:resource:scope=Namespaced,categories=all;kyma-project;gateway
// +kubebuilder:printcolumn:name="Gateway",type=string,JSONPath=`.spec.gateway`
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
