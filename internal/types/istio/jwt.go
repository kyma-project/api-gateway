package jwt

// JwtAuth Config for Jwt istio authorization
type JwtAuth struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
}

// JwtConfig is used to deserialize jwt accessStrategy configuration for istio handler
type JwtConfig struct {
	Authentications []JwtAuth `json:"authentications,omitempty"`
}
