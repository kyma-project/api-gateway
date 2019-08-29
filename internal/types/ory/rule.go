package ory

// OauthIntrospectionConfig Config for Oauth2 Oathkeeper AccessRule
type OauthIntrospectionConfig struct {
	// Array of required scopes
	RequiredScope []string `json:"required_scope"`
}
