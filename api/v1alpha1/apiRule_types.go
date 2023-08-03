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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// Status code describing APIRule.
type StatusCode string

const (
	//StatusOK .
	StatusOK StatusCode = "OK"
	//StatusSkipped .
	StatusSkipped StatusCode = "SKIPPED"
	//StatusError .
	StatusError StatusCode = "ERROR"
)

// Defines the desired state of ApiRule.
type APIRuleSpec struct {
	// Describes the service to expose.
	Service *Service `json:"service"`
	// Specifies the Istio Gateway to be used.
	// +kubebuilder:validation:Pattern=`^[0-9a-z-_]+(\/[0-9a-z-_]+|(\.[0-9a-z-_]+)*)$`
	Gateway *string `json:"gateway"`
	//Represents the array of Oathkeeper access rules to be applied.
	// +kubebuilder:validation:MinItems=1
	Rules []Rule `json:"rules"`
}

// Describes the observed state of ApiRule.
type APIRuleStatus struct {
	LastProcessedTime           *metav1.Time           `json:"lastProcessedTime,omitempty"`
	ObservedGeneration          int64                  `json:"observedGeneration,omitempty"`
	APIRuleStatus               *APIRuleResourceStatus `json:"APIRuleStatus,omitempty"`
	VirtualServiceStatus        *APIRuleResourceStatus `json:"virtualServiceStatus,omitempty"`
	AccessRuleStatus            *APIRuleResourceStatus `json:"accessRuleStatus,omitempty"`
	RequestAuthenticationStatus *APIRuleResourceStatus `json:"requestAuthenticationStatus,omitempty"`
	AuthorizationPolicyStatus   *APIRuleResourceStatus `json:"authorizationPolicyStatus,omitempty"`
}

// APIRule is the Schema for ApiRule APIs.
// +kubebuilder:deprecatedversion:warning="Since Kyma 2.5.X, APIRule in version v1alpha1 has been deprecated. Consider using v1beta1."
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type APIRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIRuleSpec   `json:"spec,omitempty"`
	Status APIRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// APIRuleList contains a list of ApiRule
type APIRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIRule `json:"items"`
}

// Service .
type Service struct {
	// Specifies the name of the exposed service.
	Name *string `json:"name"`
	// Specifies the communication port of the exposed service.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *uint32 `json:"port"`
	// Specifies the URL of the exposed service.
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:Pattern=^([a-zA-Z0-9][a-zA-Z0-9-_]*\.)*[a-zA-Z0-9]*[a-zA-Z0-9-_]*[[a-zA-Z0-9]+$
	Host *string `json:"host"`
	// Specifies if the service is internal (in cluster) or external.
	// +optional
	IsExternal *bool `json:"external,omitempty"`
}

// Rule .
type Rule struct {
	// Specifies the path of the exposed service.
	// +kubebuilder:validation:Pattern=^([0-9a-zA-Z./*()?!\\_-]+)
	Path string `json:"path"`
	// Represents the list of allowed HTTP request methods available for the **spec.rules.path**.
	// +kubebuilder:validation:MinItems=1
	Methods []string `json:"methods"`
	// Specifies the list of access strategies.
	// All strategies listed in [Oathkeeper documentation](https://www.ory.sh/docs/oathkeeper/pipeline/authn) are supported.
	// +kubebuilder:validation:MinItems=1
	AccessStrategies []*Authenticator `json:"accessStrategies"`
	// Specifies the list of [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutators.
	// +optional
	Mutators []*Mutator `json:"mutators,omitempty"`
}

// Describes the status of APIRule.
type APIRuleResourceStatus struct {
	Code        StatusCode `json:"code,omitempty"`
	Description string     `json:"desc,omitempty"`
}

func init() {
	SchemeBuilder.Register(&APIRule{}, &APIRuleList{})
}

// Represents a handler that authenticates provided credentials. See the corresponding type in the oathkeeper-maester project. provided credentials. See the corresponding type in the oathkeeper-maester project.
type Authenticator struct {
	*Handler `json:",inline"`
}

// Mutator represents a handler that transforms the HTTP request before forwarding it. See the corresponding in the oathkeeper-maester project.
type Mutator struct {
	*Handler `json:",inline"`
}

// Handler provides configuration for different Oathkeeper objects. It is used to either validate a request (Authenticator, Authorizer) or modify it (Mutator). See the corresponding type in the oathkeeper-maester project.
type Handler struct {
	// Specifies the name of the handler.
	Name string `json:"handler"`
	// Configures the handler. Configuration keys vary per handler.
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Config *runtime.RawExtension `json:"config,omitempty"`
}
