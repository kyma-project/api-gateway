package helpers

import (
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/pkg/errors"
	"net/http"
)

type RetryableHttpClient struct {
	client *http.Client
	opts   []retry.Option
}

func NewClientWithRetry(c *http.Client, opts []retry.Option) *RetryableHttpClient {
	return &RetryableHttpClient{
		client: c,
		opts:   opts,
	}
}

// CallEndpointWithRetries returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *RetryableHttpClient) CallEndpointWithRetries(url string, validator HttpResponseAsserter) error {
	err := h.withRetries(func() (*http.Response, error) {
		return h.client.Get(url)
	}, validator)

	if err != nil {
		return fmt.Errorf("error calling endpoint %s err=%s", url, err)
	}

	return nil
}

// CallEndpointWithHeadersWithRetries returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *RetryableHttpClient) CallEndpointWithHeadersWithRetries(requestHeaders map[string]string, url string, validator HttpResponseAsserter) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	for headerName, headerValue := range requestHeaders {
		req.Header.Set(headerName, headerValue)
	}

	err = h.withRetries(func() (*http.Response, error) {
		return h.client.Do(req)
	}, validator)

	if err != nil {
		return fmt.Errorf("error calling endpoint %s err=%s", url, err)
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
