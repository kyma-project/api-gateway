package v2alpha1

// OauthModeConfig Config for Oauth mode
type OauthModeConfig struct {
	// Array of paths. Each path creates an oathkeeper AccessRule
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:UniqueItems=true
	Paths []Option `json:"paths"`
}

//Option Set of options for the Oauth mode
type Option struct {
	// Path to be exposed
	// +kubebuilder:validation:Pattern=^/([0-9a-zA-Z./*]+)
	Path string `json:"path"`
	// Set of allowed Oauth scopes
	Scopes []string `json:"scopes,omitempty"`
	// Set of allowed HTTP methods
	Methods []string `json:"methods,omitempty"`
}
