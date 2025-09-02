package oauth2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type GetTokenOptions struct {
	Scopes    []string
	Format    string
	GrantType string
	Audiences []string
}

type GetTokenOption func(*GetTokenOptions)

func WithScope(scope string) GetTokenOption {
	return func(o *GetTokenOptions) {
		o.Scopes = append(o.Scopes, scope)
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
		o.Audiences = append(o.Audiences, audience)
	}
}

type RequestOption func(*RequestOptions)

type RequestOptions struct {
	GetTokenOptions []GetTokenOption
	TokenHeader     string
	TokenPrefix     string
	FromParam       string
	WithoutToken    bool
	TokenOverride   string
}

func WithTokenHeader(header string) RequestOption {
	return func(o *RequestOptions) {
		o.TokenHeader = header
	}
}

func WithTokenPrefix(prefix string) RequestOption {
	return func(o *RequestOptions) {
		o.TokenPrefix = prefix
	}
}

func WithTokenFromParam(param string) RequestOption {
	return func(o *RequestOptions) {
		o.FromParam = param
	}
}

func WithGetTokenOption(opt GetTokenOption) RequestOption {
	return func(o *RequestOptions) {
		o.GetTokenOptions = append(o.GetTokenOptions, opt)
	}
}

func WithGetTokenOptions(opt ...GetTokenOption) RequestOption {
	return func(o *RequestOptions) {
		o.GetTokenOptions = append(o.GetTokenOptions, opt...)
	}
}

func WithoutToken() RequestOption {
	return func(o *RequestOptions) {
		o.WithoutToken = true
	}
}

func WithTokenOverride(token string) RequestOption {
	return func(o *RequestOptions) {
		o.TokenOverride = token
	}
}

type Provider interface {
	GetIssuerURL() string
	GetJwksURI() string

	GetToken(t *testing.T, options ...GetTokenOption) (string, error)
	MakeRequest(t *testing.T, method, url string, options ...RequestOption) (statusCode int, responseHeaders map[string][]string, responseBody []byte, err error)
}

// AssertEndpointWithProvider asserts that the given endpoint responds correctly with
// a token from the provided OAuth2 provider.
// It checks that requests:
//   - without a token return response code 403,
//   - with an invalid token return response code 401,
//   - with a valid token return response code 200.
func AssertEndpointWithProvider(t *testing.T, provider Provider, url string, method string, options ...RequestOption) {
	t.Helper()

	statusCode, _, _, err := provider.MakeRequest(t, method, url, append(options, WithoutToken())...)
	assert.NoError(t, err)
	assert.Equal(t, 403, statusCode)

	statusCode, _, _, err = provider.MakeRequest(t, method, url, append(options, WithTokenOverride("not.good.token"))...)
	assert.NoError(t, err)
	assert.Equal(t, 401, statusCode)

	statusCode, _, _, err = provider.MakeRequest(t, method, url, options...)
	assert.NoError(t, err)
	assert.Equal(t, 200, statusCode)
}
