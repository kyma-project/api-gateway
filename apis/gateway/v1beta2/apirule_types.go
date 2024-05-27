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

package v1beta2

import (
	"istio.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type State string

const (
	Ready      State = "Ready"
	Processing State = "Processing"
	Error      State = "Error"
	Deleting   State = "Deleting"
	Warning    State = "Warning"
)

// APIRuleSpec defines the desired state of ApiRule.
type APIRuleSpec struct {
	// Specifies the URLs of the exposed service.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=1
	Hosts []*Host `json:"hosts"`
	// Describes the service to expose.
	// +optional
	Service *Service `json:"service,omitempty"`
	// Specifies the Istio Gateway to be used.
	// +kubebuilder:validation:Pattern=`^[0-9a-z-_]+(\/[0-9a-z-_]+|(\.[0-9a-z-_]+)*)$`
	Gateway *string `json:"gateway"`
	// Specifies CORS headers configuration that will be sent downstream
	// +optional
	CorsPolicy *CorsPolicy `json:"corsPolicy,omitempty"`
	// Represents the array of Oathkeeper access rules to be applied.
	// +kubebuilder:validation:MinItems=1
	Rules []Rule `json:"rules"`
	// +optional
	Timeout *Timeout `json:"timeout,omitempty"`
}

// Host is the URL of the exposed service.
// +kubebuilder:validation:MinLength=3
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=^([a-zA-Z0-9][a-zA-Z0-9-_]*\.)*[a-zA-Z0-9]*[a-zA-Z0-9-_]*[[a-zA-Z0-9]+$
type Host string

// APIRuleStatus describes the observed state of ApiRule.
type APIRuleStatus struct {
	LastProcessedTime *metav1.Time `json:"lastProcessedTime,omitempty"`
	// State signifies current state of APIRule.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting", "Warning").
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state"`
	// Description of APIRule status
	Description string `json:"description,omitempty"`
}

// APIRule is the Schema for ApiRule APIs.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Hosts",type="string",JSONPath=".spec.hosts"
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
	// Specifies the Namespace of the exposed service. If not defined, it defaults to the APIRule Namespace.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// Specifies the communication port of the exposed service.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *uint32 `json:"port"`
	// Specifies if the service is internal (in cluster) or external.
	// +optional
	IsExternal *bool `json:"external,omitempty"`
}

// Rule .
// +kubebuilder:validation:XValidation:rule="has(self.jwt) ? !has(self.noAuth) || self.noAuth == false : has(self.noAuth) && self.noAuth == true",message="either jwt is configured or noAuth must be set to true in a rule"
type Rule struct {
	// Specifies the path of the exposed service.
	// +kubebuilder:validation:Pattern=^([0-9a-zA-Z./*()?!\\_-]+)
	Path string `json:"path"`
	// Describes the service to expose. Overwrites the **spec** level service if defined.
	// +optional
	Service *Service `json:"service,omitempty"`
	// Represents the list of allowed HTTP request methods available for the **spec.rules.path**.
	// +kubebuilder:validation:MinItems=1
	Methods []HttpMethod `json:"methods"`
	// Disables authorization when set to true.
	// +optional
	NoAuth *bool `json:"noAuth"`
	// Specifies the Istio JWT access strategy.
	// +optional
	Jwt *JwtConfig `json:"jwt,omitempty"`
	// +optional
	Timeout *Timeout `json:"timeout,omitempty"`
}

// HttpMethod specifies the HTTP request method. The list of supported methods is defined in RFC 9910: HTTP Semantics and RFC 5789: PATCH Method for HTTP.
// +kubebuilder:validation:Enum=GET;HEAD;POST;PUT;DELETE;CONNECT;OPTIONS;TRACE;PATCH
type HttpMethod string

func init() {
	SchemeBuilder.Register(&APIRule{}, &APIRuleList{})
}

// JwtConfig is the configuration for the Istio JWT authentication
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type JwtConfig struct {
	Authentications []*JwtAuthentication `json:"authentications,omitempty"`
	Authorizations  []*JwtAuthorization  `json:"authorizations,omitempty"`
}

func (j *JwtConfig) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// JwtAuthorization contains an array of required scopes
type JwtAuthorization struct {
	RequiredScopes []string `json:"requiredScopes"`
	Audiences      []string `json:"audiences"`
}

// JwtAuthentication Config for Jwt Istio authentication
type JwtAuthentication struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
	// +optional
	FromHeaders []*JwtHeader `json:"fromHeaders,omitempty"`
	// +optional
	FromParams []string `json:"fromParams,omitempty"`
}

// JwtHeader for specifying from header for the Jwt token
type JwtHeader struct {
	Name string `json:"name"`
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// Timeout for HTTP requests in seconds. The timeout can be configured up to 3900 seconds (65 minutes).
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=3900
type Timeout uint16 // We use unit16 instead of a time.Duration because there is a bug with duration that requires additional validation of the format. Issue: checking https://github.com/kubernetes/apiextensions-apiserver/issues/56

const (
	Regex  = "regex"
	Exact  = "exact"
	Prefix = "prefix"
)

type StringMatch []map[string]string

func (s StringMatch) ToIstioStringMatchArray() (out []*v1beta1.StringMatch) {
	for _, match := range s {
		for key, value := range match {
			switch key {
			case Regex:
				out = append(out, &v1beta1.StringMatch{MatchType: &v1beta1.StringMatch_Regex{Regex: value}})
			case Exact:
				out = append(out, &v1beta1.StringMatch{MatchType: &v1beta1.StringMatch_Exact{Exact: value}})
			case Prefix:
				fallthrough
			default:
				out = append(out, &v1beta1.StringMatch{MatchType: &v1beta1.StringMatch_Prefix{Prefix: value}})
			}
		}
	}
	return out
}

// CorsPolicy allows configuration of CORS headers received downstream. If this is not defined, the default values are applied.
// If CorsPolicy is configured, CORS headers received downstream will be only those defined on the APIRule
type CorsPolicy struct {
	AllowHeaders     []string    `json:"allowHeaders,omitempty"`
	AllowMethods     []string    `json:"allowMethods,omitempty"`
	AllowOrigins     StringMatch `json:"allowOrigins,omitempty"`
	AllowCredentials *bool       `json:"allowCredentials,omitempty"`
	ExposeHeaders    []string    `json:"exposeHeaders,omitempty"`
        // +kubebuilder:validation:Minimum=0
	MaxAge           uint64      `json:"maxAge,omitempty"`
}
