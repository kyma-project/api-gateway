/*

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

package v2alpha1

import (
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//StatusCode .
type StatusCode string

const (
	//StatusOK .
	StatusOK StatusCode = "OK"
	//StatusSkipped .
	StatusSkipped StatusCode = "SKIPPED"
	//StatusError .
	StatusError StatusCode = "ERROR"
)

// GateSpec defines the desired state of Gate
type GateSpec struct {
	// Definition of the service to expose
	Service *Service `json:"service"`
	// Gateway to be used
	// +kubebuilder:validation:Pattern=^(?:[_a-z0-9](?:[_a-z0-9-]+[a-z0-9])?\.)+(?:[a-z](?:[a-z0-9-]+[a-z0-9])?)?$
	Gateway *string `json:"gateway"`
	//Paths represents collection of Path to secure
	// +kubebuilder:validation:MinItems=1
	Rules []Rule `json:"rules,omitempty"`
}

// GateStatus defines the observed state of Gate
type GateStatus struct {
	LastProcessedTime    *metav1.Time           `json:"lastProcessedTime,omitempty"`
	ObservedGeneration   int64                  `json:"observedGeneration,omitempty"`
	GateStatus           *GatewayResourceStatus `json:"GateStatus,omitempty"`
	VirtualServiceStatus *GatewayResourceStatus `json:"virtualServiceStatus,omitempty"`
	AccessRuleStatus     *GatewayResourceStatus `json:"accessRuleStatus,omitempty"`
}

//Gate is the Schema for the apis Gate
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Gate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GateSpec   `json:"spec,omitempty"`
	Status GateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GateList contains a list of Gate
type GateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gate `json:"items"`
}

//Service .
type Service struct {
	// Name of the service
	Name *string `json:"name"`
	// Port of the service to expose
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *uint32 `json:"port"`
	// URL on which the service will be visible
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:Pattern=^(?:[_a-z0-9](?:[_a-z0-9-]+[a-z0-9])?\.)+(?:[a-z](?:[a-z0-9-]+[a-z0-9])?)?$
	Host *string `json:"host"`
	// Defines if the service is internal (in cluster) or external
	// +optional
	IsExternal *bool `json:"external,omitempty"`
}

//Rule .
type Rule struct {
	// Path to be exposed
	// +kubebuilder:validation:Pattern=^/([0-9a-zA-Z./*]+)
	Path string `json:"path"`
	// Set of allowed HTTP methods
	Methods []string `json:"methods,omitempty"`
	// Set of access strategies for a single path
	// +kubebuilder:validation:MinItems=1
	AccessStrategies []*rulev1alpha1.Authenticator `json:"accessStrategies"`
	// Mutators to be used
	// +optional
	Mutators []*rulev1alpha1.Mutator `json:"mutators,omitempty"`
}

//GatewayResourceStatus .
type GatewayResourceStatus struct {
	Code        StatusCode `json:"code,omitempty"`
	Description string     `json:"desc,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Gate{}, &GateList{})
}
