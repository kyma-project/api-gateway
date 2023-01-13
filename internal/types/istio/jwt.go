package jwt

// JwtAuthentication Config for Jwt istio authentication
type JwtAuthentication struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
}

// JwtAuthorization config for Jwt istio authorization
type JwtAuthorization struct {
	RequiredScopes []string `json:"requiredScopes"`
}

// JwtConfig is used to deserialize jwt accessStrategy configuration for istio handler
type JwtConfig struct {
	Authentications []JwtAuthentication `json:"authentications,omitempty"`
	Authorizations  []JwtAuthorization  `json:"authorizations,omitempty"`
}
