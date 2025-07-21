package oauth2mock

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"io"
	"net/http"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"testing"
	"text/template"
)

//go:embed manifest.yaml
var rawManifest string

type Mock struct {
	IssuerURL                 string
	TokenURL                  string
	JwksURI                   string
	VirtualServiceDestination string
	Subdomain                 string

	parsedManifest []byte
}

type Options struct {
	Namespace string
	Domain    string
}

func WithNamespace(ns string) Option {
	return func(o *Options) {
		o.Namespace = ns
	}
}

func WithDomain(domain string) Option {
	return func(o *Options) {
		o.Domain = domain
	}
}

type Option func(*Options)

func DeployMock(t *testing.T, options ...Option) (*Mock, error) {
	t.Helper()
	opts := &Options{
		Namespace: "oauth2-mock",
		Domain:    "local.kyma.dev",
	}
	for _, opt := range options {
		opt(opts)
	}

	mock := &Mock{
		IssuerURL:                 fmt.Sprintf("http://mock-oauth2-server.%s.svc.cluster.local", opts.Namespace),
		VirtualServiceDestination: fmt.Sprintf("mock-oauth2-server.%s.svc.cluster.local", opts.Namespace),
		JwksURI:                   fmt.Sprintf("http://mock-oauth2-server.%s.svc.cluster.local/oauth2/certs", opts.Namespace),
		TokenURL:                  fmt.Sprintf("https://%s.%s/oauth2/token", opts.Namespace, opts.Domain),
		Subdomain:                 fmt.Sprintf("%s.%s", opts.Namespace, opts.Domain),
	}

	t.Logf("Deploying oauth2mock with IssuerURL: %s, TokenURL: %s, Subdomain: %s",
		mock.IssuerURL, mock.TokenURL, mock.Subdomain)
	return mock, startMock(t, mock, opts)
}

func startMock(t *testing.T, m *Mock, options *Options) error {
	t.Helper()
	r, err := infrahelpers.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	err = infrahelpers.CreateNamespace(t, options.Namespace, infrahelpers.IgnoreAlreadyExists())
	if err != nil {
		t.Logf("Failed to create namespace: %v", err)
		return fmt.Errorf("failed to create namespace %s: %w", options.Namespace, err)
	}

	// No further cleanup is needed as the namespace will be deleted
	// as part of Namespace cleanup.
	// setup.DeclareCleanup(t, func() {})

	return m.start(t, r, options)
}

func (m *Mock) start(t *testing.T, r *resources.Resources, options *Options) error {
	err := m.parseTmpl()
	if err != nil {
		return err
	}

	err = decoder.DecodeEach(
		t.Context(),
		bytes.NewBuffer(m.parsedManifest),
		decoder.CreateHandler(r),
		decoder.MutateNamespace(options.Namespace),
	)
	if err != nil {
		t.Logf("Failed to deploy mock: %v", err)
		return err
	}

	return wait.For(conditions.New(r).DeploymentAvailable("mock-oauth2-server-deployment", options.Namespace))
}

func (m *Mock) parseTmpl() error {
	var sbuf bytes.Buffer
	tmpl, err := template.New("").Parse(rawManifest)
	if err != nil {
		return err
	}
	err = tmpl.Execute(&sbuf, m)
	if err != nil {
		return err
	}
	m.parsedManifest = sbuf.Bytes()
	return nil
}

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

type tokenStruct struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (m *Mock) GetToken(t *testing.T, options ...GetTokenOption) (string, error) {
	t.Helper()
	opts := &GetTokenOptions{
		Format:    "jwt", // Default format is JWT
		GrantType: "client_credentials",
	}
	for _, opt := range options {
		opt(opts)
	}

	t.Logf("Getting token with options: %+v", opts)

	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("mock-token-client"))
	requestBody := fmt.Sprintf("grant_type=%s&token_format=%s", opts.GrantType, opts.Format)
	if opts.Audience != "" {
		requestBody += fmt.Sprintf("&audience=%s", opts.Audience)
	}
	if opts.Scope != "" {
		requestBody += fmt.Sprintf("&scope=%s", opts.Scope)
	}

	request, err := http.NewRequest(http.MethodPost, m.TokenURL, bytes.NewBufferString(requestBody))
	if err != nil {
		t.Logf("Failed to create request: %v", err)
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(request)
	if err != nil {
		t.Logf("Failed to get token: %v", err)
		return "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			t.Logf("Failed to read response body: %v", readErr)
			return "", fmt.Errorf("failed to get token, status code: %d, error reading body: %w", resp.StatusCode, readErr)
		}
		t.Logf("Failed to get token, status code: %d, response body: %s", resp.StatusCode, body)
		return "", fmt.Errorf("failed to get token, status code: %d", resp.StatusCode)
	}

	var token tokenStruct
	jsonBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Logf("Failed to read response body: %v", readErr)
		return "", fmt.Errorf("failed to read response body: %w", readErr)
	}
	err = json.Unmarshal(jsonBody, &token)
	if err != nil {
		t.Logf("Failed to unmarshal token response: %v", err)
		return "", fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	if token.AccessToken == "" {
		t.Logf("Failed to get token, access_token is empty")
		return "", fmt.Errorf("failed to get token, access_token is empty")
	}
	t.Logf("Successfully got token: %s", token.AccessToken)
	return token.AccessToken, nil
}

func (m *Mock) MakeRequestWithMockToken(t *testing.T, method, url string, options ...GetTokenOption) (statusCode int, responseHeaders map[string][]string, responseBody []byte, err error) {
	t.Helper()
	token, err := m.GetToken(t, options...)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get token: %w", err)
	}

	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("mock-JWT-client"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := httpClient.Do(request)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to make request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Logf("Failed to close response body: %v", closeErr)
		}
	}()
	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}
	responseHeaders = make(map[string][]string)
	for key, values := range resp.Header {
		responseHeaders[key] = values // Take the first value for simplicity
	}

	statusCode = resp.StatusCode
	t.Logf("Request to %s returned status code %d with headers %v and body %s", url, statusCode, responseHeaders, responseBody)
	return statusCode, responseHeaders, responseBody, nil
}
