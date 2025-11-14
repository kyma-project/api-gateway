package endpoint

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/stretchr/testify/assert"
)

func AssertEndpoint(t *testing.T, method, url string, expectedHttpCode int) error {
	t.Helper()
	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("http-client-go"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")

	return nil
}

func AssertEndpointWithoutResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedMissingHeaders []string) error {
	t.Helper()
	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("http-client-go"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for headerName, headerValue := range requestHeaders {
		request.Header.Set(headerName, headerValue)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")

	if len(expectedMissingHeaders) > 0 {
		for _, header := range expectedMissingHeaders {
			assert.Empty(t, response.Header.Get(header))
		}
	}

	return nil
}

func AssertEndpointWithResponseHeaders(t *testing.T, method, url string, requestHeaders map[string]string, expectedHttpCode int, expectedResponseHeaders map[string]string) error {
	t.Helper()
	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("http-client-go"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for headerName, headerValue := range requestHeaders {
		request.Header.Set(headerName, headerValue)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")
	for headerName, headerValue := range expectedResponseHeaders {
		responseHeaderValue := response.Header.Get(headerName)
		if headerValue != responseHeaderValue {
			t.Fatalf("Didn't get the expected response header: %s: %s, got %s", headerName, headerValue, responseHeaderValue)
		}
	}

	return nil
}
