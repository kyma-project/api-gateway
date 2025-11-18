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
	"github.com/kyma-project/api-gateway/apis/gateway/versions"
	"istio.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type State string
// Defines the reconciliation state of the APIRule. 
// The possible states are Ready, Warning, Error, Processing, or Deleting.

const (
	// The APIRule's reconciliation is finished.
	Ready      State = "Ready"
	// The APIRule is being created or updated.
	Processing State = "Processing"
	// An error occurred during reconciliation.
	Error      State = "Error"
	// The APIRule is being deleted.
	Deleting   State = "Deleting"
	// The APIRule is misconfigured.
	Warning    State = "Warning"
)

// **APIRuleSpec** defines the desired state of the APIRule.
type APIRuleSpec struct {
	// Specifies the Service’s communication address for inbound external traffic. 
	// The following formats are supported:
	// - A fully qualified domain name (FQDN) with at least two domain labels separated by dots. Each label must consist of lowercase alphanumeric characters or '-', 
	// and must start and end with a lowercase alphanumeric character. For example, `my-example.domain.com`, or `example.com`.
	// - One lowercase RFC 1123 label (referred to as short host name) that must consist of lowercase alphanumeric characters or '-', and must start and end with a lowercase alphanumeric character. For example, `my-host`.
	// If you define a single label, the domain name is taken from the Gateway referenced in the APIRule. In this case, the Gateway must provide the same single host for all Server definitions 
	// and it must be prefixed with `*.`. Otherwise, the validation fails.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=1
	Hosts []*Host `json:"hosts"`
	// Specifies the backend Service that receives traffic. The Service can be deployed inside the cluster.
	// If you don't define a Service at the **spec.service** level, each defined rule must 
	// specify a Service at the **spec.rules.service** level. Otherwise, the validation fails.
	// +optional
	Service *Service `json:"service,omitempty"`
	// Specifies the Istio Gateway. The field must reference an existing Gateway in the cluster. 
	// Provide the Gateway in the format `namespace/gateway`. 
	// Both the namespace and the Gateway name cannot be longer than 63 characters each.
	// +kubebuilder:validation:MaxLength=127
	// +kubebuilder:validation:XValidation:rule=`self.matches('^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?/([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)$')`,message="Gateway must be in the namespace/name format"
	Gateway *string `json:"gateway"`
	// Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the CORS headers are removed from the response.
	// +optional
	CorsPolicy *CorsPolicy `json:"corsPolicy,omitempty"`
	/* Defines an ordered list of access rules. Each rule is an atomic configuration that 
	defines how to access a specific HTTP path. A rule consists of a path 
	pattern, one or more allowed HTTP methods, exactly one access strategy (**jwt**, **extAuth**, 
	or **noAuth**), and other optional configuration fields. */
	// +kubebuilder:validation:MinItems=1
	Rules []Rule `json:"rules"`
	// Specifies the timeout for HTTP requests in seconds for all rules. 
	// You can override the value for each rule. If no timeout is specified, the default timeout of 180 seconds applies.
	// +optional
	Timeout *Timeout `json:"timeout,omitempty"`
}

// The host is the URL of the exposed Service. Lowercase RFC 1123 labels and FQDN are supported.
// +kubebuilder:validation:MaxLength=255
// +kubebuilder:validation:XValidation:rule=`self.matches('^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:(?:\\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*(?:\\.[a-z0-9]{2,63}))?$')`,message="Host must be a lowercase RFC 1123 label (must consist of lowercase alphanumeric characters or '-', and must start and end with an lowercase alphanumeric character) or a fully qualified domain name"
type Host string

// Describes the observed status of the APIRule.
type APIRuleStatus struct {
	// Represents the last time the APIRule status was processed.
	LastProcessedTime metav1.Time `json:"lastProcessedTime,omitempty"`
	// Defines the reconciliation state of the APIRule. 
	// The possible states are `Ready`, `Warning`, or `Error`.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state"`
	// Contains the description of the APIRule's status.
	Description string `json:"description,omitempty"`
}

func (s *APIRuleStatus) ApiRuleStatusVersion() versions.Version {
	return versions.V2
}

// APIRule is the schema for APIRule APIs.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-api-gateway}
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="Hosts",type="string",JSONPath=".spec.hosts"
type APIRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Defines the desired state of the APIRule.
	// +kubebuilder:validation:Required
	Spec   APIRuleSpec   `json:"spec"`
	// Describes the observed status of the APIRule.
	Status APIRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// APIRuleList contains a list of APIRules
type APIRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIRule `json:"items"`
}

	// Specifies the backend Service that receives traffic. The Service must be deployed inside the cluster.
	// If you don't define a Service at the **spec.service** level, each defined rule must 
	// specify a Service at the **spec.rules.service** level. Otherwise, the validation fails.
type Service struct {
	// Specifies the name of the exposed Service.
	Name *string `json:"name"`
	// Specifies the namespace of the exposed Service.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +optional
	Namespace *string `json:"namespace,omitempty"`
	// Specifies the communication port of the exposed Service.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *uint32 `json:"port"`
	// Specifies if the Service is internal (deployed in the cluster) or external.
	// +optional
	IsExternal *bool `json:"external,omitempty"`
}

// Defines an ordered list of access rules. Each rule is an atomic access configuration that 
// defines how to access a specific HTTP path. A rule consists of a path pattern, one or more 
// allowed HTTP methods, exactly one access strategy (`jwt`, `extAuth`, or `noAuth`), 
// and other optional configuration fields. The order of rules in the APIRule CR is important. 
// Rules defined earlier in the list have a higher priority than those defined later.
// +kubebuilder:validation:XValidation:rule="((has(self.extAuth)?1:0)+(has(self.jwt)?1:0)+((has(self.noAuth)&&self.noAuth==true)?1:0))==1",message="One of the following fields must be set: noAuth, jwt, extAuth"
type Rule struct {
	// Specifies the path on which the Service is exposed. The supported configurations are:
	//  - Exact path (e.g. /abc) - matches the specified path exactly.
	//  - The `{*}` operator (for example, `/foo/{*}` or `/foo/{*}/bar`) - matches 
	// any request that matches the pattern with exactly one path segment in the operator's place.
	//  - The `{**}` operator (for example, `/foo/{**}` or `/foo/{**}/bar`) -
	//  matches any request that matches the pattern with zero or more path segments in the operator's place.
	//  The `{**}` operator must be the last operator in the path.
	//  - The wildcard path `/*` - matches all paths. Equivalent to the `/{**}` path.
	// The value might contain the operators `{*}` and/or `{**}`. It can also be a wildcard match `/*`.
	// For more information, see [Ordering Rules in APIRule v2](https://kyma-project.io/external-content/api-gateway/docs/user/custom-resources/apirule/04-20-significance-of-rule-path-and-method-order.html).
	// +kubebuilder:validation:Pattern=`^((\/([A-Za-z0-9-._~!$&'()+,;=:@]|%[0-9a-fA-F]{2})*)|(\/\{\*{1,2}\}))+$|^\/\*$`
	Path string `json:"path"`
	// Specifies the backend Service that receives traffic. The Service must be deployed inside the cluster.
	// If you don't define a Service at the **spec.service** level, each defined rule must
	// specify a Service at the **spec.rules.service** level. Otherwise, the validation fails.
	// +optional
	Service *Service `json:"service,omitempty"`
	// Specifies the list of HTTP request methods available for spec.rules.path. 
	// The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) 
	// and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).
	// +kubebuilder:validation:MinItems=1
	Methods []HttpMethod `json:"methods"`
	// Disables authorization when set to `true`.
	// +optional
	NoAuth *bool `json:"noAuth"`
	// Specifies the Istio JWT configuration.
	// +optional
	Jwt *JwtConfig `json:"jwt,omitempty"`
	// Specifies the external authorization configuration.
	// +optional
	ExtAuth *ExtAuth `json:"extAuth,omitempty"`
	// Specifies the timeout, in seconds, for HTTP requests made to spec.rules.path. 
	// Timeout definitions set at this level take precedence over any timeout defined 
	// at the spec.timeout level. The maximum timeout is limited to 3900 seconds (65 minutes).
	// +optional
	Timeout *Timeout `json:"timeout,omitempty"`
	// Defines request modification rules, which are applied before forwarding the request to the target workload.
	// +optional
	Request *Request `json:"request,omitempty"`
}

type Request struct {
	// Specifies a list of cookie key-value pairs, that are forwarded inside the Cookie header.
	// +optional
	Cookies map[string]string `json:"cookies,omitempty"`
	// Specifies a list of header key-value pairs that are forwarded as header=value to the target workload.
	// +optional
	Headers map[string]string `json:"headers,omitempty"`
}

// HttpMethod specifies the HTTP request method. The list of supported methods is defined in in 
// [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).
// +kubebuilder:validation:Enum=GET;HEAD;POST;PUT;DELETE;CONNECT;OPTIONS;TRACE;PATCH
type HttpMethod string

func init() {
	SchemeBuilder.Register(&APIRule{}, &APIRuleList{})
}

// Configures Istio JWT authentication and authorization.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type JwtConfig struct {
	// Specifies the list of authentication objects.
	Authentications []*JwtAuthentication `json:"authentications,omitempty"`
	// Specifies the list of authorization objects.
	Authorizations  []*JwtAuthorization  `json:"authorizations,omitempty"`
}

func (j *JwtConfig) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// Specifies the list of Istio JWT authorization objects.
type JwtAuthorization struct {
	// Specifies the list of required scope values for the JWT.
	// +optional
	RequiredScopes []string `json:"requiredScopes,omitempty"`
	// Specifies the list of audiences required for the JWT.
	// +optional
	Audiences []string `json:"audiences,omitempty"`
}

// Specifies the list of Istio JWT authentication objects.
type JwtAuthentication struct {
	// Identifies the issuer that issued the JWT. The value must be a URL. 
	// Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.
	Issuer  string `json:"issuer"`
	// Contains the URL of the provider’s public key set to validate the signature of the JWT. 
	// The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.
	JwksUri string `json:"jwksUri"`
	// Specifies the list of headers from which the JWT token is extracted.
	// +optional
	FromHeaders []*JwtHeader `json:"fromHeaders,omitempty"`
	// Specifies the list of parameters from which the JWT token is extracted.
	// +optional
	FromParams []string `json:"fromParams,omitempty"`
}

// Specifies the header from which the JWT token is extracted. 
type JwtHeader struct {
	// Specifies the name of the header from which the JWT token is extracted.
	Name string `json:"name"`
	// Specifies the prefix used before the JWT token. The default is `Bearer`.
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// **ExtAuth** contains configuration for paths that use external authorization.
type ExtAuth struct {
	// Specifies the name of the external authorization handler.
	// +kubebuilder:validation:MinItems=1
	ExternalAuthorizers []string `json:"authorizers"`
	// Specifies JWT configuration for the external authorization handler.
	// +optional
	Restrictions *JwtConfig `json:"restrictions,omitempty"`
}

// Specifies the timeout for HTTP requests in seconds for all rules. 
// You can override the value for each rule. If no timeout is specified, the default timeout of 180 seconds applies.
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

// Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, 
// the CORS headers are removed from the response.
type CorsPolicy struct {
	// Indicates whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.
	AllowHeaders     []string    `json:"allowHeaders,omitempty"`
	// Lists headers allowed with the **Access-Control-Allow-Headers** CORS header.
	AllowMethods     []string    `json:"allowMethods,omitempty"`
	// Lists headers allowed with the **Access-Control-Allow-Methods** CORS header.
	AllowOrigins     StringMatch `json:"allowOrigins,omitempty"`
	// Lists origins allowed with the **Access-Control-Allow-Origins** CORS header.
	AllowCredentials *bool       `json:"allowCredentials,omitempty"`
	// Lists headers allowed with the **Access-Control-Expose-Headers** CORS header.
	ExposeHeaders    []string    `json:"exposeHeaders,omitempty"`
	// Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header.
	// +kubebuilder:validation:Minimum=1
	MaxAge *uint64 `json:"maxAge,omitempty"`
}
