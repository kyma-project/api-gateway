package mock

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/dnswait"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/setup/ipfamily"
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
	Domain string
}

func WithDomain(domain string) Option {
	return func(o *Options) {
		o.Domain = domain
	}
}

type Option func(*Options)

func (m *Mock) GetIssuerURL() string {
	return m.IssuerURL
}

func (m *Mock) GetJwksURI() string {
	return m.JwksURI
}

func DeployMock(t *testing.T, ns string, options ...Option) (*Mock, error) {
	t.Helper()
	opts := &Options{
		Domain: "local.kyma.dev",
	}
	for _, opt := range options {
		opt(opts)
	}

	mock := &Mock{
		IssuerURL:                 fmt.Sprintf("http://mock-oauth2-server.%s.svc.cluster.local", ns),
		VirtualServiceDestination: fmt.Sprintf("mock-oauth2-server.%s.svc.cluster.local", ns),
		JwksURI:                   fmt.Sprintf("http://mock-oauth2-server.%s.svc.cluster.local/oauth2/certs", ns),
		TokenURL:                  fmt.Sprintf("https://%s.%s/oauth2/token", ns, opts.Domain),
		Subdomain:                 fmt.Sprintf("%s.%s", ns, opts.Domain),
	}

	t.Logf("Deploying oauth2mock with GetIssuerURL: %s, TokenURL: %s, Subdomain: %s",
		mock.IssuerURL, mock.TokenURL, mock.Subdomain)
	return mock, startMock(t, ns, mock, opts)
}

func startMock(t *testing.T, ns string, m *Mock, options *Options) error {
	t.Helper()
	r, err := client.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return err
	}

	return m.start(t, ns, r, options)
}

func (m *Mock) start(t *testing.T, ns string, r *resources.Resources, options *Options) error {
	err := m.parseTmpl()
	if err != nil {
		return err
	}

	err = decoder.DecodeEach(
		t.Context(),
		bytes.NewBuffer(m.parsedManifest),
		decoder.CreateHandler(r),
		decoder.MutateNamespace(ns),
		decoder.MutateOption(mutateServiceIPFamily),
	)
	if err != nil {
		t.Logf("Failed to deploy mock: %v", err)
		return err
	}

	setup.DeclareCleanup(t, func() {
		t.Logf("Cleaning up oauth2mock in namespace %s", ns)
		err := decoder.DecodeEach(
			setup.GetCleanupContext(),
			bytes.NewBuffer(m.parsedManifest),
			decoder.DeleteHandler(r),
			decoder.MutateNamespace(ns),
		)
		if err != nil {
			t.Logf("Failed to clean up oauth2mock: %v", err)
		} else {
			t.Logf("Successfully cleaned up oauth2mock in namespace %s", ns)
		}
	})

	return wait.For(conditions.New(r).DeploymentAvailable("mock-oauth2-server-deployment", ns))
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

type tokenStruct struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (m *Mock) GetToken(t *testing.T, options ...oauth2.GetTokenOption) (string, error) {
	t.Helper()
	// Wait for the mock's own TokenURL to resolve externally. On Gardener
	// AWS this hostname is served by the same DNSEntry wildcard as the
	// APIRule host; propagation lags in-cluster Ready by 30-90s. Skipped
	// on k3d and for IP-literal URLs (see dnswait.WaitForURL).
	dnswait.WaitForURL(t, m.TokenURL)
	opts := &oauth2.GetTokenOptions{
		Format:    "jwt", // Default format is JWT
		GrantType: "client_credentials",
	}
	for _, opt := range options {
		opt(opts)
	}

	t.Logf("Getting token with options: %+v", opts)

	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("mock-token-client"))
	requestBody := fmt.Sprintf("grant_type=%s&token_format=%s", opts.GrantType, opts.Format)
	if len(opts.Audiences) > 0 {
		requestBody += fmt.Sprintf("&audience=%s", strings.Join(opts.Audiences, ","))
	}
	if len(opts.Scopes) > 0 {
		requestBody += fmt.Sprintf("&scope=%s", strings.Join(opts.Scopes, " "))
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

func (m *Mock) MakeRequest(t *testing.T, method, url string, options ...oauth2.RequestOption) (statusCode int, responseHeaders map[string][]string, responseBody []byte, err error) {
	t.Helper()
	// Wait for the caller-supplied APIRule host to resolve externally.
	// The GetToken() call below waits separately on the mock's TokenURL.
	dnswait.WaitForURL(t, url)
	opts := &oauth2.RequestOptions{
		TokenHeader: "Authorization",
		TokenPrefix: "Bearer",
	}
	for _, opt := range options {
		opt(opts)
	}

	token := opts.TokenOverride
	if token == "" {
		// ponytail: token fetch uses resolver-default dialling, only the
		// authenticated request below is family-pinned. The dualstack gate
		// runs on a shoot with both families, so token retrieval resolves
		// regardless; the per-family assertion is on the authenticated dial.
		t, err := m.GetToken(t, opts.GetTokenOptions...)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("failed to get token: %w", err)
		}
		token = t
	}

	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("mock-JWT-client"), httphelper.WithNetwork(opts.Network))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if opts.WithHeaders != nil {
		for headerName, headerValue := range opts.WithHeaders {
			request.Header.Set(headerName, headerValue)
		}
	}

	if opts.FromParam == "" && !opts.WithoutToken {
		request.Header.Set(opts.TokenHeader, fmt.Sprintf("%s %s", opts.TokenPrefix, token))
	} else if !opts.WithoutToken {
		request.URL.RawQuery = fmt.Sprintf("%s=%s", opts.FromParam, token)
	}

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
		responseHeaders[key] = values
	}

	statusCode = resp.StatusCode
	t.Logf("Request to %s returned status code %d with headers %v and body %s", url, statusCode, responseHeaders, responseBody)
	return statusCode, responseHeaders, responseBody, nil
}

// mutateServiceIPFamily aligns each core/v1 Service in the mock manifest
// with the TEST_IP_FAMILY axis via ipfamily.ApplyToService. Keeps the
// OAuth2 mock reachable in-cluster over v6 in dualstack mode and behaves
// as SingleStack IPv4 in the unset / ipv4 modes.
func mutateServiceIPFamily(obj k8s.Object) error {
	if svc, ok := obj.(*corev1.Service); ok {
		ipfamily.ApplyToService(svc)
	}
	return nil
}
