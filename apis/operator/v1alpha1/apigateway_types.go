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
	// APIGateway Controller finished reconciliation.
	Ready      State = "Ready"
	// APIGateway Controller is reconciling resources.
	Processing State = "Processing"
	// An error occurred during the reconciliation. 
	// The error is rather related to the API Gateway module than the configuration of your resources.
	Error      State = "Error"
	// APIGateway Controller is deleting resources.
	Deleting   State = "Deleting"
	// An issue occurred during reconciliation that requires your attention.
	// Check the **status.description** message to identify the issue and make the necessary corrections 
	// to the APIGateway CR or any related resources.
	Warning    State = "Warning"
)

// Defines the desired state of APIGateway CR.
type APIGatewaySpec struct {

	// Specifies whether the default Kyma Gateway `kyma-gateway` in `kyma-system` namespace is created.
	// +optional
	EnableKymaGateway *bool `json:"enableKymaGateway,omitempty"`
}

// Defines the observed state of APIGateway CR.
type APIGatewayStatus struct {
	// State signifies current state of APIGateway. The possible values are `Ready`, `Processing`, `Error`, `Deleting`, `Warning`.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state"`
	// Contains the description of the APIGateway's state.
	Description string `json:"description,omitempty"`
	// Contains conditions associated with the APIGateway's status.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".status.state",name="State",type="string"
//+kubebuilder:resource:scope=Cluster,categories={kyma-modules,kyma-api-gateway}

// APIGateway is the Schema for APIGateway APIs.
type APIGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of APIGateway CR.
	Spec   APIGatewaySpec   `json:"spec,omitempty"`
	// Defines the observed status of APIGateway CR.
	Status APIGatewayStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// APIGatewayList contains a list of APIGateways.
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

func GetOldestAPIGatewayCR(apigatewayCRs *APIGatewayList) *APIGateway {
	if len(apigatewayCRs.Items) == 0 {
		return nil
	}

	oldest := apigatewayCRs.Items[0]
	for _, item := range apigatewayCRs.Items {
		timestamp := &item.CreationTimestamp
		if !(oldest.CreationTimestamp.Before(timestamp)) {
			oldest = item
		}
	}

	return &oldest
}
