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
	// The APIRule reconciliation is finished.
	StatusReady   = "Ready"
	// The RateLimit is misconfigured.
	StatusWarning = "Warning"
	// An error occurred during reconciliation.
	StatusError   = "Error"
)

// Contains rate limit bucket configuration.
// +kubebuilder:validation:XValidation:rule="has(self.path) || has(self.headers)",message="At least one of 'path' or 'headers' must be set"
type BucketConfig struct {
	// Specifies the path for which rate limiting is applied. The path must start with `/`. For example, `/foo`.
	Path    string            `json:"path,omitempty"`
	// Specifies the headers for which rate limiting is applied. The key is the header's name, and the value is the header's value. 
	// All specified headers must be present in the request for this configuration to match. For example, `x-api-usage: BASIC`.
	Headers map[string]string `json:"headers,omitempty"`
	// Defines the token bucket specification.
	// +kubebuilder:validation:Required
	Bucket BucketSpec `json:"bucket"`
}

// Defines the token bucket specification.
type BucketSpec struct {
	// The maximum number of tokens that the bucket can hold. 
	// This is also the number of tokens that the bucket initially contains.
	// +kubebuilder:validation:Required
	MaxTokens int64 `json:"maxTokens"`
	// The number of tokens added to the bucket during each fill interval.
	// +kubebuilder:validation:Required
	TokensPerFill int64 `json:"tokensPerFill"`
	// Specifies the fill interval. During each fill interval, the number of tokens specified in the 
	// **tokensPerFill** field is added to the bucket. 
	// The bucket cannot contain more than maxTokens tokens.
	// The fillInterval must be greater than or equal to 50ms to avoid excessive refills.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=duration
	FillInterval *metav1.Duration `json:"fillInterval"`
}

// Defines the local rate limit configuration.
type LocalConfig struct {
	// The default token bucket for rate limiting requests.
	// If additional local buckets are configured in the same RateLimit CR, this bucket serves as a fallback for requests that don't match any other bucket's criteria.
	// Each request consumes a single token. If a token is available, the request is allowed. If no tokens are available, the request is rejected with status code `429`.
	// +kubebuilder:validation:Required
	DefaultBucket BucketSpec     `json:"defaultBucket"`
	// Specifies a list of additional rate limit buckets for requests. Each bucket must specify either a path or headers.
	// For each request matching the bucket's criteria, a single token is consumed. If a token is available, the request is allowed. 
	// If no tokens are available, the request is rejected with status code `429`.
	Buckets []BucketConfig `json:"buckets,omitempty"`
}

// Defines the desired state of the RateLimit custom resource.
type RateLimitSpec struct {
	// Contains labels that specify the set of Pods or `istio-ingressgateway` to which the configuration applies.
	// Each Pod must match only one RateLimit CR.
	// The label scope is limited to the namespace where the resource is located.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinProperties=1
	SelectorLabels map[string]string `json:"selectorLabels"`
	// Defines the local rate limit configuration.
	// +kubebuilder:validation:Required
	Local LocalConfig `json:"local"`
	// Enables **x-rate-limit** response headers. The default value is `false`.
	EnableResponseHeaders bool `json:"enableResponseHeaders,omitempty"`
	// Controls whether rate limiting is enforced. If true, requests exceeding limits are rejected. 
	// If false, request limits are monitored but requests that exceed limits are not blocked. 
	// The default value is `true`.
	//+kubebuilder:default:=true
	Enforce bool `json:"enforce,omitempty"`
}

// RateLimitStatus defines the observed state of RateLimit
type RateLimitStatus struct {
	// Description defines the description of current State of RateLimit.
	Description string `json:"description,omitempty"`
	// State describes the overall status of RateLimit. The possible values are `Ready`, `Warning`, and `Error`.
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

func (s *RateLimitStatus) Warning(err error) {
	s.State = StatusWarning
	s.Description = err.Error()
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RateLimit is the Schema for reate limits API.
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RateLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Defines the desired state of the RateLimit custom resource.
	Spec   RateLimitSpec   `json:"spec,omitempty"`
	// Defines the current state of the RateLimit custom resource.
	Status RateLimitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// RateLimitList contains a list of RateLimit custom resources.
type RateLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RateLimit{}, &RateLimitList{})
}
