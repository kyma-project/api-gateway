package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"
)

type RetryableHttpClient struct {
	client *http.Client
	opts   []retry.Option
}

type RequestOptions struct {
	Audiences []string
	Scopes    []string
}

func NewClientWithRetry(c *http.Client, opts []retry.Option) *RetryableHttpClient {
	return &RetryableHttpClient{
		client: c,
		opts:   opts,
	}
}

func (h *RetryableHttpClient) CallEndpointWithRetriesAndGetResponse(headers map[string]string, body io.Reader, method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	var resp *http.Response
	err = retry.Do(func() error {
		resp, err = h.client.Do(req)
		if err != nil {
			return err
		}
		return nil
	}, h.opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CallEndpointWithRetries returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *RetryableHttpClient) CallEndpointWithRetries(url string, validator HttpResponseAsserter) error {
	err := h.withRetries(func() (*http.Response, error) {
		return h.client.Get(url)
	}, validator)

	if err != nil {
		return fmt.Errorf("timestamp: %s -- error calling endpoint %s err=%s", time.Now().String(), url, err)
	}

	return nil
}

// CallEndpointWithHeadersWithRetries calls given url with headers and GET method. Returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *RetryableHttpClient) CallEndpointWithHeadersWithRetries(requestHeaders map[string]string, url string, validator HttpResponseAsserter) error {
	return h.CallEndpointWithHeadersAndMethod(requestHeaders, url, http.MethodGet, validator)
}

// CallEndpointWithHeadersAndMethod calls given url with given method and headers. Returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *RetryableHttpClient) CallEndpointWithHeadersAndMethod(requestHeaders map[string]string, url string, method string, validator HttpResponseAsserter) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	for headerName, headerValue := range requestHeaders {
		req.Header.Set(headerName, headerValue)
	}

	return h.CallEndpointWithRequestRetry(req, validator)
}

func (h *RetryableHttpClient) CallEndpointWithRequestRetry(req *http.Request, validator HttpResponseAsserter) error {
	err := h.withRetries(func() (*http.Response, error) {
		return h.client.Do(req)
	}, validator)

	if err != nil {
		return fmt.Errorf("error calling endpoint %s err=%s", req.URL, err)
	}

	return nil
}

func (h *RetryableHttpClient) withRetries(httpCall func() (*http.Response, error), validator HttpResponseAsserter) error {

	if err := retry.Do(func() error {

		response, callErr := httpCall()
		if callErr != nil {
			return callErr
		}

		if isValid, failureMsg := validator.Assert(*response); !isValid {
			return errors.New(failureMsg)
		}

		return nil
	},
		h.opts...,
	); err != nil {
		return err
	}

	return nil
}

type OIDCConfiguration struct {
	Issuer                                string   `json:"issuer"`
	AuthorizationEndpoint                 string   `json:"authorization_endpoint"`
	TokenEndpoint                         string   `json:"token_endpoint"`
	UserinfoEndpoint                      string   `json:"userinfo_endpoint"`
	EndSessionEndpoint                    string   `json:"end_session_endpoint"`
	JwksUri                               string   `json:"jwks_uri"`
	IntrospectionEndpoint                 string   `json:"introspection_endpoint"`
	RevocationEndpoint                    string   `json:"revocation_endpoint"`
	ResponseTypesSupported                []string `json:"response_types_supported"`
	GrantTypesSupported                   []string `json:"grant_types_supported"`
	SubjectTypesSupported                 []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported      []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                       []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported     []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                       []string `json:"claims_supported"`
	CodeChallengeMethodsSupported         []string `json:"code_challenge_methods_supported"`
	TlsClientCertificateBoundAccessTokens bool     `json:"tls_client_certificate_bound_access_tokens"`
	FrontchannelLogoutSupported           bool     `json:"frontchannel_logout_supported"`
	FrontchannelLogoutSessionSupported    bool     `json:"frontchannel_logout_session_supported"`
}

func GetOIDCConfiguration(oidcConfigurationEndpoint string) (oidcConfiguration OIDCConfiguration, err error) {
	var resp *http.Response
	err = retry.Do(func() error {
		response, err := http.Get(oidcConfigurationEndpoint)
		resp = response
		return err
	}, retry.Attempts(20), retry.Delay(2*time.Second), retry.DelayType(retry.FixedDelay))

	if err != nil {
		return OIDCConfiguration{}, err
	}
	err = json.NewDecoder(resp.Body).Decode(&oidcConfiguration)
	if err != nil {
		return OIDCConfiguration{}, err
	}
	return oidcConfiguration, err
}
