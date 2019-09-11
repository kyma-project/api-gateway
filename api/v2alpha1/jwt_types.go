package v2alpha1

// JWTModeConfig config for JWT mode
type JWTModeConfig struct {
	Issuer string   `json:"issuer"`
	JWKS   []string `json:"jwks,omitempty"`
}

// JWTModeALL representation of config for the ALL mode
type JWTModeALL struct {
	Scopes []string `json:"scopes"`
}

// JWTModeInclude representation of config for the INCLUDE mode
type JWTModeInclude struct {
	Paths []IncludePath `json:"paths"`
}

// IncludePath Path for INCLUDE mode
type IncludePath struct {
	Path    string   `json:"path"`
	Scopes  []string `json:"scopes"`
	Methods []string `json:"methods"`
}
