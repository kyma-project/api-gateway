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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type StatusCode string

const (
	JWT            string     = "JWT"
	OAUTH          string     = "OAUTH"
	PASSTHROUGH    string     = "PASSTHROUGH"
	STATUS_OK      StatusCode = "OK"
	STATUS_SKIPPED StatusCode = "SKIPPED"
	STATUS_ERROR   StatusCode = "ERROR"
)

// ApiSpec defines the desired state of Api
type ApiSpec struct {
	// Definition of the service to expose
	Service *Service `json:"service"`
	// Auth strategy to be used
	Auth *AuthStrategy `json:"auth"`
}

// ApiStatus defines the observed state of Api
type ApiStatus struct {
	LastProcessedTime    *metav1.Time           `json:"lastProcessedTime,omitempty"`
	ObservedGeneration   int64                  `json:"observedGeneration,omitempty"`
	APIStatus            *GatewayResourceStatus `json:"APIStatus,omitempty"`
	VirtualServiceStatus *GatewayResourceStatus `json:"virtualServiceStatus,omitempty"`
	PolicyServiceStatus  *GatewayResourceStatus `json:"policyStatus,omitempty"`
	AccessRuleStatus     *GatewayResourceStatus `json:"accessRuleStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Api is the Schema for the apis API
type Api struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiSpec   `json:"spec,omitempty"`
	Status ApiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApiList contains a list of Api
type ApiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Api `json:"items"`
}

type Service struct {
	// Name of the service
	Name *string `json:"name"`
	// Port of the service to expose
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=99999
	Port *int32 `json:"port"`
	// URL on which the service will be visible
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:Pattern=^(?:https?:\/\/)?(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)
	HostURL *string `json:"hostURL"`
	// Defines if the service is internal (in cluster) or external
	// +optional
	IsExternal *bool `json:"external,omitempty"`
}

type AuthStrategy struct {
	// +kubebuilder:validation:Enum=JWT;OAUTH;PASSTHROUGH
	Name   *string               `json:"name"`
	Config *runtime.RawExtension `json:"config,inline"`
}

type GatewayResourceStatus struct {
	Code        StatusCode `json:"code,omitempty"`
	Description string     `json:"desc,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Api{}, &ApiList{})
}
