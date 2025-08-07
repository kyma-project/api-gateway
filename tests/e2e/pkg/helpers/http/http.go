package httphelper

import (
	"crypto/tls"
	"net/http"
	"testing"
)

type Options struct {
	Prefix string
}

type Option func(*Options)

func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
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

	transport := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{
		Transport: TestLogTransportWrapper(t, opts.Prefix, &transport),
	}
}

type RoundTripFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestLogTransportWrapper(t *testing.T, prefix string, rt http.RoundTripper) RoundTripFunc {
	return func(req *http.Request) (*http.Response, error) {
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
