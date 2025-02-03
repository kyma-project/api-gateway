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

package v2

import (
	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories={kyma-api-gateway}
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="Hosts",type="string",JSONPath=".spec.hosts"

// APIRule is the Schema for the apirules API
type APIRule v2alpha1.APIRule

//+kubebuilder:object:root=true

// APIRuleList contains a list of APIRule
type APIRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIRule `json:"items"`
}

type APIRuleSpec v2alpha1.APIRuleSpec

type APIRuleStatus v2alpha1.APIRuleStatus

func init() {
	SchemeBuilder.Register(&APIRule{}, &APIRuleList{})
}
