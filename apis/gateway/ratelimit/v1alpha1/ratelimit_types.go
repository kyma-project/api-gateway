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

const (
	StatusReady = "Ready"
	StatusError = "Error"
)

// BucketConfig represents a rate limit bucket configuration.
// +kubebuilder:validation:XValidation:rule="((has(self.path)?1:0)+(has(self.headers)?1:0))==1",message="path or headers must be set"
type BucketConfig struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	// +kubebuilder:validation:Required
	Bucket BucketSpec `json:"bucket"`
}

// BucketSpec defines the token bucket specification.
type BucketSpec struct {
	// +kubebuilder:validation:Required
	MaxTokens int64 `json:"maxTokens"`
	// +kubebuilder:validation:Required
	TokensPerFill int64 `json:"tokensPerFill"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=duration
	FillInterval *metav1.Duration `json:"fillInterval"`
}

// LocalConfig represents the local rate limit configuration.
type LocalConfig struct {
	// +kubebuilder:validation:Required
	DefaultBucket BucketSpec     `json:"defaultBucket"`
	Buckets       []BucketConfig `json:"buckets,omitempty"`
}

// RateLimitSpec defines the desired state of RateLimit
type RateLimitSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1
	SelectorLabels map[string]string `json:"selectorLabels"`
	// +kubebuilder:validation:Required
	Local LocalConfig `json:"local"`
	// EnableResponseHeaders enables x-rate-limit response headers. The default value is false.
	EnableResponseHeaders bool `json:"enableResponseHeaders,omitempty"`
	// Enforce specifies whether the rate limit should be enforced. The default value is `true`.
	//+kubebuilder:default:=true
	Enforce bool `json:"enforce,omitempty"`
}

// RateLimitStatus defines the observed state of RateLimit
type RateLimitStatus struct {
	// Description defines the description of current State of RateLimit.
	Description string `json:"description,omitempty"`
	// State describes the overall status of RateLimit. Values are `Ready`, `Processing` and `Error`
	State string `json:"state,omitempty"`
}

func (s *RateLimitStatus) Error(err error) {
	s.State = StatusError
	s.Description = err.Error()
}

func (s *RateLimitStatus) Ready() {
	s.State = StatusReady
	s.Description = "Finished reconciliation"
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RateLimit is the Schema for the ratelimits API
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
