package httphelper

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"testing"
	"time"
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

// TestLogTransportWrapper wraps rt with request/response logging and
// applies the configured Host header + extra headers to every request.
// The host and headers arguments may be empty; they are applied only when
// non-empty.
func TestLogTransportWrapper(t *testing.T, prefix, host string, headers map[string]string, rt http.RoundTripper) RoundTripFunc {
	return func(req *http.Request) (*http.Response, error) {
		if host != "" {
			req.Host = host
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		t.Logf("[%s] request method: %s, url: %s", prefix, req.Method, req.URL)
		t.Logf("[%s] request headers: %v", prefix, req.Header)

		resp, err := rt.RoundTrip(req)
		if err != nil {
			t.Logf("[%s] request error: method: %s, url: %s, err: %v", prefix, req.Method, req.URL, err)
			return nil, err
		}
		t.Logf("[%s] response: %d %s", prefix, resp.StatusCode, http.StatusText(resp.StatusCode))
		return resp, nil
	}
}
