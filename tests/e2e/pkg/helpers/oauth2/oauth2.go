package oauth2

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/dnswait"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
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
	WithHeaders     map[string]string
	Network         string
}

// WithNetwork pins the TCP family ("tcp4"/"tcp6", or "" for resolver
// default) used by the provider's HTTP client for this request. Callers
// obtain the value from ipfamily.ForEachDialNetwork so the authenticated
// dial is exercised over the family selected by TEST_IP_FAMILY.
func WithNetwork(network string) RequestOption {
	return func(o *RequestOptions) {
		o.Network = network
	}
}

func WithHeaders(headers map[string]string) RequestOption {
	return func(o *RequestOptions) {
		if o.WithHeaders == nil {
			o.WithHeaders = make(map[string]string)
		}
		for k, v := range headers {
			o.WithHeaders[k] = v
		}
	}
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
//
// The provider owns its HTTP client construction, so we cannot hand it the
// family-pinned client from ForEachDialNetwork. Instead we use the fan-out
// only for its per-family t.Run wrapping + network selection and thread the
// network into the provider via WithNetwork. Single-family mode (ipv4/unset)
// runs once with network "" — byte-identical to before.
func AssertEndpointWithProvider(t *testing.T, provider Provider, url string, method string, options ...RequestOption) {
	t.Helper()
	// Same pre-dial DNS wait that endpoint.AssertEndpoint uses. The JWT
	// path builds its own HTTP client through the oauth2 Provider, so it
	// bypasses runPerFamilyOnURL — we call dnswait.WaitForURL here to
	// close the same Route 53 propagation gap.
	dnswait.WaitForURL(t, url)

	ipfamily.ForEachDialNetwork(t, "oauth2-provider", nil, func(t *testing.T, network string, _ *http.Client) {
		opts := append(options, WithNetwork(network))

		statusCode, _, _, err := provider.MakeRequest(t, method, url, append(opts, WithoutToken())...)
		assert.NoError(t, err)
		assert.Equal(t, 403, statusCode)

		statusCode, _, _, err = provider.MakeRequest(t, method, url, append(opts, WithTokenOverride("not.good.token"))...)
		assert.NoError(t, err)
		assert.Equal(t, 401, statusCode)

		statusCode, _, _, err = provider.MakeRequest(t, method, url, opts...)
		assert.NoError(t, err)
		assert.Equal(t, 200, statusCode)
	})
}

// AssertNonExposedEndpointWithProvider asserts that the given not exposed endpoint responds correctly with 404
func AssertNonExposedEndpointWithProvider(t *testing.T, provider Provider, url string, method string, options ...RequestOption) {
	t.Helper()
	dnswait.WaitForURL(t, url)

	statusCode, _, _, err := provider.MakeRequest(t, method, url, append(options, WithoutToken())...)
	assert.NoError(t, err)
	assert.Equal(t, 404, statusCode)
}
