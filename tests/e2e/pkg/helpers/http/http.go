package httphelper

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/artifacts"
)

// Options configures NewHTTPClient. All fields have zero-value fallbacks
// so a bare NewHTTPClient(t) still works.
//
// Network pins the socket family used for outgoing dials. Valid values are
// "tcp4", "tcp6", or empty (leave the choice to the resolver). Pair with
// ipfamily.From().DialNetworks() when a test must exercise both families
// on a dualstack cluster.
type Options struct {
	Prefix  string
	Host    string
	Headers map[string]string
	Timeout time.Duration
	Network string
}

type Option func(*Options)

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

// WithHost sets the HTTP Host header (req.Host) on outgoing requests. Useful
// when a test dials the LB by IP or by a wildcard hostname but needs a
// specific Host value for HTTP routing.
func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

// WithHeaders adds request headers applied to every outgoing request.
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		o.Headers = headers
	}
}

// WithTimeout sets http.Client.Timeout. Zero or negative values keep the
// default (no timeout).
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithNetwork pins the TCP family for outgoing dials. Values: "tcp4",
// "tcp6", or empty for resolver-default. When set, the client's Transport
// uses a custom DialContext that ignores the caller-supplied network and
// dials with this family instead.
func WithNetwork(network string) Option {
	return func(o *Options) {
		o.Network = network
	}
}

func NewHTTPClient(t *testing.T, options ...Option) *http.Client {
	t.Helper()
	opts := &Options{
		Prefix: "http-test-client",
	}
	for _, opt := range options {
		opt(opts)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if opts.Network != "" {
		dialer := &net.Dialer{Timeout: 30 * time.Second}
		transport.DialContext = func(ctx context.Context, _, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, opts.Network, addr)
		}
	}
	client := &http.Client{
		Transport: TestLogTransportWrapper(t, opts.Prefix, opts.Host, opts.Headers, transport),
	}
	if opts.Timeout > 0 {
		client.Timeout = opts.Timeout
	}
	return client
}

type RoundTripFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type TestLogTransportWrapperOptions struct {
	SuppressTestLog bool
	Output          io.Writer
	outputMu        sync.Mutex
}

type TestLogTransportOption func(*TestLogTransportWrapperOptions)

func SuppressTestLog() TestLogTransportOption {
	return func(o *TestLogTransportWrapperOptions) {
		o.SuppressTestLog = true
	}
}

func WithOutput(output io.Writer) TestLogTransportOption {
	return func(o *TestLogTransportWrapperOptions) {
		o.Output = output
	}
}

func logfWithOptions(t *testing.T, prefix string, opts *TestLogTransportWrapperOptions, format string, args ...interface{}) {
	sbuilder := &strings.Builder{}
	sbuilder.WriteString(fmt.Sprintf("[%s] ", prefix))
	sbuilder.WriteString(fmt.Sprintf(format, args...))
	toLog := sbuilder.String()

	if !opts.SuppressTestLog {
		t.Log(toLog)
	}
	if opts.Output != nil {
		opts.outputMu.Lock()
		_, err := io.WriteString(opts.Output, toLog+"\n")
		opts.outputMu.Unlock()
		if err != nil {
			t.Logf("Warning: failed to write to output: %v", err)
		}
	}
}

func TestLogTransportWrapper(t *testing.T, prefix string, host string, headers map[string]string, rt http.RoundTripper, option ...TestLogTransportOption) RoundTripFunc {
	opts := &TestLogTransportWrapperOptions{}
	for _, opt := range option {
		opt(opts)
	}

	return func(req *http.Request) (*http.Response, error) {
		// Set Host header if specified
		if host != "" {
			req.Host = host
		}

		logfWithOptions(t, prefix, opts, "request Host header set to: %s", req.Host)

		// Set custom headers if specified
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		logfWithOptions(t, prefix, opts, "request method: %s, url: %s, host: %s", req.Method, req.URL, req.Host)
		logfWithOptions(t, prefix, opts, "request headers: %v", req.Header)

		resp, err := rt.RoundTrip(req)
		if err != nil {
			logfWithOptions(t, prefix, opts, "request failed; method: %s, url: %s, err: %v", req.Method, req.URL, err)
			return nil, err
		}
		logfWithOptions(t, prefix, opts, "received response; status code: %d, status text: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return resp, nil
	}
}

const (
	httpLogsDir = "http-logs"
)

func OpenTestArtifactLog(t *testing.T, name string) io.Writer {
	t.Helper()

	dir := filepath.Join(artifacts.Root(), artifacts.TestRunTimestamp(), artifacts.SanitizePathComponent(t.Name()), httpLogsDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Logf("Warning: failed to create artifact dir %s: %v", dir, err)
		return nil
	}

	filePath := filepath.Join(dir, artifacts.SanitizePathComponent(name)+".log")
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Logf("Warning: failed to open artifact log %s: %v", filePath, err)
		return nil
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Logf("Warning: failed to close artifact log %s: %v", filePath, err)
		}
	})
	return f
}
