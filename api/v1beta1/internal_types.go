package v1beta1

//JWTAccStrConfig is used to deserialize jwt accessStrategy configuration for the validation purposes
type JWTAccStrConfig struct {
	TrustedIssuers []string `json:"trusted_issuers,omitempty"`
	JWKSUrls       []string `json:"jwks_urls,omitempty"`
	RequiredScopes []string `json:"required_scopes,omitempty"`
}
