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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type State string

const (
	Ready      State = "Ready"
	Processing State = "Processing"
	Error      State = "Error"
	Deleting   State = "Deleting"
	Warning    State = "Warning"
)

// APIGatewaySpec defines the desired state of APIGateway
type APIGatewaySpec struct {

	// Specifies whether the default Kyma Gateway kyma-gateway in kyma-system Namespace is created.
	// +optional
	EnableKymaGateway *bool `json:"enableKymaGateway,omitempty"`
}

// APIGatewayStatus defines the observed state of APIGateway
type APIGatewayStatus struct {
	// State signifies current state of APIGateway. Value can be one of ("Ready", "Processing", "Error", "Deleting", "Warning").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state"`
	// Description of APIGateway status
	Description string `json:"description,omitempty"`
	// Conditions of APIGateway
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.state",name="State",type="string"
//+kubebuilder:resource:scope=Cluster,categories={kyma,kyma-modules,kyma-api-gateway}

// APIGateway is the Schema for the apigateways API
type APIGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIGatewaySpec   `json:"spec,omitempty"`
	Status APIGatewayStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// APIGatewayList contains a list of APIGateway
type APIGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIGateway{}, &APIGatewayList{})
}

// IsInDeletion returns true if the APIGateway is in deletion process, false otherwise.
func (a *APIGateway) IsInDeletion() bool {
	return !a.DeletionTimestamp.IsZero()
}

func (a *APIGateway) HasFinalizer() bool {
	return len(a.Finalizers) > 0
}
