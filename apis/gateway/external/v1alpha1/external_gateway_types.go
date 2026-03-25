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
	// ExternalDomain is the customer-facing domain (e.g., api.customer.com or *.api.customer.com)
	// Supports Istio Gateway host format with optional wildcard prefix (e.g., *.example.com)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Pattern=`^(\*\.)?([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`
	ExternalDomain string `json:"externalDomain"`

	// InternalDomain configuration for Kyma-internal access
	// +kubebuilder:validation:Required
	InternalDomain InternalDomainConfig `json:"internalDomain"`

	// BTPRegion is a BTP region identifier (e.g., "eu10", "us10")
	// This must match a btp_region defined in the external-gateway-regions ConfigMap
	// +kubebuilder:validation:Required
	BTPRegion string `json:"btpRegion"`

	// RegionsConfigMap is the name of the ConfigMap containing UGW region metadata.
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

// GatewayName returns the generated name for the Istio Gateway resource
// Format: {externalgateway-name}-gateway
func (e *ExternalGateway) GatewayName() string {
	return e.Name + "-gateway"
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
