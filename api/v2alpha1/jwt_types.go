package v2alpha1

import "k8s.io/apimachinery/pkg/runtime"

// JWTModeConfig config for JWT mode
type JWTModeConfig struct {
	Issuer string         `json:"issuer"`
	JWKS   []string       `json:"jwks,omitempty"`
	Mode   InternalConfig `json:"mode"`
}

// InternalConfig internal config, specific for JWT modes
type InternalConfig struct {
	Name   string                `json:"name"`
	Config *runtime.RawExtension `json:"config,omitempty"`
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

const (
	// JWTAll ?
	JWTAll string = "ALL"
	// JWTInclude ?
	JWTInclude string = "INCLUDE"
	// JWTExclude ?
	JWTExclude string = "EXCLUDE"
)
