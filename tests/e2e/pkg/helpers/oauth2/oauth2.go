package oauth2

import "testing"

type GetTokenOptions struct {
	Scope     string
	Format    string
	GrantType string
	Audience  string
}

type GetTokenOption func(*GetTokenOptions)

func WithScope(scope string) GetTokenOption {
	return func(o *GetTokenOptions) {
		o.Scope = scope
	}
}

func WithOpaqueTokenFormat() GetTokenOption {
	return func(o *GetTokenOptions) {
		o.Format = "opaque"
	}
}

func WithJWTTokenFormat() GetTokenOption {
	return func(o *GetTokenOptions) {
		o.Format = "jwt"
	}
}

func WithAudience(audience string) GetTokenOption {
	return func(o *GetTokenOptions) {
		o.Audience = audience
	}
}

type Provider interface {
	GetIssuerURL() string
	GetJwksURI() string

	GetToken(t *testing.T, options ...GetTokenOption) (string, error)
	MakeRequestWithToken(t *testing.T, method, url string, options ...GetTokenOption) (statusCode int, responseHeaders map[string][]string, responseBody []byte, err error)
}
