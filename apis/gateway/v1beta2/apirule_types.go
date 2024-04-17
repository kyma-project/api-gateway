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
	"k8s.io/apimachinery/pkg/runtime"
)

// StatusCode describing APIRule.
type StatusCode string

const (
	//StatusOK is set when the reconciliation finished successfully
	StatusOK StatusCode = "OK"
	//StatusSkipped is set when reconciliation of the APIRule component was skipped
	StatusSkipped StatusCode = "SKIPPED"
	//StatusError is set when an error happened during reconciliation of the APIRule
	StatusError StatusCode = "ERROR"
	//StatusWarning is set if a user action is required
	StatusWarning StatusCode = "WARNING"
)

// APIRuleSpec defines the desired state of ApiRule.
type APIRuleSpec struct {
	// Specifies the URL of the exposed service.
	Hosts []*string `json:"hosts"`
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

// APIRuleStatus describes the observed state of ApiRule.
type APIRuleStatus struct {
	LastProcessedTime    *metav1.Time           `json:"lastProcessedTime,omitempty"`
	ObservedGeneration   int64                  `json:"observedGeneration,omitempty"`
	APIRuleStatus        *APIRuleResourceStatus `json:"APIRuleStatus,omitempty"`
	VirtualServiceStatus *APIRuleResourceStatus `json:"virtualServiceStatus,omitempty"`
	// +optional
	AccessRuleStatus *APIRuleResourceStatus `json:"accessRuleStatus,omitempty"`
	// +optional
	RequestAuthenticationStatus *APIRuleResourceStatus `json:"requestAuthenticationStatus,omitempty"`
	// +optional
	AuthorizationPolicyStatus *APIRuleResourceStatus `json:"authorizationPolicyStatus,omitempty"`
}

// APIRule is the Schema for ApiRule APIs.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.APIRuleStatus.code"
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".spec.host"
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
// +kubebuilder:pruning:PreserveUnknownFields
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
	// Specifies the list of external authorizers.
	// +optional
	ExtAuths []*ExtAuth `json:"extAuths,omitempty"`
	// Specifies the Istio JWT access strategy.
	// +optional
	Jwt *JwtConfig `json:"jwt,omitempty"`
	// Specifies the list of [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutators.
	// +optional
	Mutators []*Mutator `json:"mutators,omitempty"`
	// +optional
	Timeout *Timeout `json:"timeout,omitempty"`
}

// HttpMethod specifies the HTTP request method. The list of supported methods is defined in RFC 9910: HTTP Semantics and RFC 5789: PATCH Method for HTTP.
// +kubebuilder:validation:Enum=GET;HEAD;POST;PUT;DELETE;CONNECT;OPTIONS;TRACE;PATCH
type HttpMethod string

type ExtAuth struct {
	// Specifies the name of the external authorizer.
	Name string `json:"name"`
}

// APIRuleResourceStatus describes the status of APIRule.
type APIRuleResourceStatus struct {
	Code        StatusCode `json:"code,omitempty"`
	Description string     `json:"desc,omitempty"`
}

func init() {
	SchemeBuilder.Register(&APIRule{}, &APIRuleList{})
}

// Mutator is a configuration for Istio mutators. It is used to enrich an incoming request with information.
type Mutator struct {
	// Specifies the name of the mutator.
	Handler string `json:"handler"`
	// Configures the mutator. Configuration keys vary per mutator.
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Config *runtime.RawExtension `json:"config,omitempty"`
}

// JwtConfig is the configuration for the Istio JWT authentication
type JwtConfig struct {
	Authentications []*JwtAuthentication `json:"authentications,omitempty"`
	Authorizations  []*JwtAuthorization  `json:"authorizations,omitempty"`
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
	// +kubebuilder:validation:Format=duration
	MaxAge *metav1.Duration `json:"maxAge,omitempty"`
}