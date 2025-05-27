package ory

// OauthIntrospectionConfig Config for Oauth2 Oathkeeper AccessRule.
type OauthIntrospectionConfig struct {
	// Array of required scopes
	RequiredScope []string `json:"required_scope"`
}

// JwtConfig Config for Jwt Oathkeeper AccessRule.
type JwtConfig struct {
	// Array of required scopes
	RequiredScope []string `json:"required_scope"`
	TrustedIssuer []string `json:"trusted_issuers"`
}
