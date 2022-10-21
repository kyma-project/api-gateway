package istio

// JwtAuth Config for Jwt istio authorization
type JwtAuth struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
}
