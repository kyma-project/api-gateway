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
	"time"
)

// +kubebuilder:validation:XValidation:rule="((has(self.path)?1:0)+(has(self.headers)?1:0))==1",message="path or headers must be set"
type Bucket struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	// +kubebuilder:validation:Required
	MaxTokens int64 `json:"maxTokens"`
	// +kubebuilder:validation:Required
	TokensPerFill int64 `json:"tokensPerFill"`
	// +kubebuilder:validation:Requiredg
	FillInterval time.Duration `json:"fillInterval"`
}

type DefaultBucket struct {
	// +kubebuilder:validation:Required
	MaxTokens int64 `json:"maxTokens"`
	// +kubebuilder:validation:Required
	TokensPerFill int64 `json:"tokensPerFill"`
	// +kubebuilder:validation:Required
	FillInterval time.Duration `json:"fillInterval"`
}

type Local struct {
	// +kubebuilder:validation:Required
	DefaultBucket DefaultBucket `json:"defaultBucket"`
	Buckets       []Bucket      `json:"buckets,omitempty"`
}

// RateLimitSpec defines the desired state of RateLimit
type RateLimitSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1
	SelectorLabels map[string]string `json:"selectorLabels"`
	// +kubebuilder:validation:Required
	Local                 Local `json:"local"`
	EnableResponseHeaders bool  `json:"enableResponseHeaders,omitempty"`
	Enforce               bool  `json:"enforce,omitempty"`
}

// RateLimitStatus defines the observed state of RateLimit
type RateLimitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RateLimit is the Schema for the ratelimits API
type RateLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RateLimitSpec   `json:"spec,omitempty"`
	Status RateLimitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RateLimitList contains a list of RateLimit
type RateLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RateLimit{}, &RateLimitList{})
}

// check if any workload exists with the given selector
// then we take all the selectors from the workload
// then we check if for those other selectors other ratelimits exists (if any RateLimit CR has a selector that is on the workload
// if yes then download other selectors, if no warning state something like there is no workload with the given selector
// check if for other selectors other

// First RateLimit CR: selectorLabels: app=boo
// Second RateLimit CR: selectorLabels: app=boo app=poo -> got into the warning state because EF is already applied to both workloads

// 1 pod: boo, poo -> First RL CR
// 2 pod: boo -> First RL CR

// Read selectors from RateLimit CR
// Check if any workload exists with the given selector
// Take all the workloads with the given selectors -> might have zoo + check if the workload is ingress-gateway or has sidecar injected
// Check if any other already existing RateLimit CRs applies to those workloads
