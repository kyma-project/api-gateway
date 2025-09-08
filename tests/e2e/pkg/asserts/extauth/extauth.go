package extauth

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httphelper "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/http"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/oauth2"
)

func AssertEndpoint(t *testing.T, method, url string, headers map[string]string, expectedHttpCode int) error {
	t.Helper()
	httpClient := httphelper.NewHTTPClient(t, httphelper.WithPrefix("ext-auth-client"))
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for headerName, headerValue := range headers {
		request.Header.Set(headerName, headerValue)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)
	assert.Equal(t, expectedHttpCode, response.StatusCode, "unexpected status code")

	return nil
}

func AssertEndpointWithJWT(t *testing.T, method, url string, expectedHttpCode int, provider oauth2.Provider, options ...oauth2.RequestOption) error {
	t.Helper()

	statusCode, _, _, err := provider.MakeRequest(t, method, url, options...)
	require.NoError(t, err, "failed to make request with JWT")
	assert.Equal(t, expectedHttpCode, statusCode, "unexpected status code")
	return nil
}
