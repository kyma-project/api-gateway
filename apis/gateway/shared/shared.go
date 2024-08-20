package shared

import "k8s.io/apimachinery/pkg/runtime/schema"

// JwtConfig is the configuration for the Istio JWT authentication and authorization.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type JwtConfig struct {
	Authentications []*JwtAuthentication `json:"authentications,omitempty"`
	Authorizations  []*JwtAuthorization  `json:"authorizations,omitempty"`
}

func (j *JwtConfig) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// JwtAuthorization contains scopes and audiences required for the JWT token.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type JwtAuthorization struct {
	// +optional
	RequiredScopes []string `json:"requiredScopes,omitempty"`
	// +optional
	Audiences []string `json:"audiences,omitempty"`
}

func (j *JwtAuthorization) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// JwtAuthentication Config for Jwt Istio authentication
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type JwtAuthentication struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
	// +optional
	FromHeaders []*JwtHeader `json:"fromHeaders,omitempty"`
	// +optional
	FromParams []string `json:"fromParams,omitempty"`
}

func (j *JwtAuthentication) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// JwtHeader for specifying from header for the Jwt token
type JwtHeader struct {
	Name string `json:"name"`
	// +optional
	Prefix string `json:"prefix,omitempty"`
}
