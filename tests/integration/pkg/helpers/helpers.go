package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"net/http"
)

// RetriableApiRule wraps any function that modifies or creates an APIRule
type RetriableApiRule func(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error)

type Helper struct {
	client *http.Client
	opts   []retry.Option
}

func NewHelper(c *http.Client, opts []retry.Option) *Helper {
	return &Helper{
		client: c,
		opts:   opts,
	}
}

// CallEndpointWithRetries returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *Helper) CallEndpointWithRetries(url string, validator HttpResponseAsserter) error {
	err := h.withRetries(func() (*http.Response, error) {
		return h.client.Get(url)
	}, validator)

	if err != nil {
		return fmt.Errorf("error calling endpoint %s err=%s", url, err)
	}

	return nil
}

// CallEndpointWithHeadersWithRetries returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *Helper) CallEndpointWithHeadersWithRetries(headerValue string, headerName, url string, validators HttpResponseAsserter) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set(headerName, headerValue)
	err = h.withRetries(func() (*http.Response, error) {
		return h.client.Do(req)
	}, validators)

	if err != nil {
		return fmt.Errorf("error calling endpoint %s err=%s", url, err)
	}

	return nil
}

func (h *Helper) withRetries(httpCall func() (*http.Response, error), validator HttpResponseAsserter) error {

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

// APIRuleWithRetries tries toExecute function and retries with onRetry if APIRule status is "ERROR"
func (h *Helper) APIRuleWithRetries(toExecute RetriableApiRule, onRetry RetriableApiRule, k8sClient dynamic.Interface, resources []unstructured.Unstructured) error {
	type status struct {
		Status struct {
			APIRuleStatus struct {
				Code string `json:"code"`
			} `json:"APIRuleStatus"`
		} `json:"status"`
	}
	res, err := toExecute(k8sClient, resources...)
	if err != nil {
		return err
	}

	js, err := json.Marshal(res)
	if err != nil {
		return err
	}

	apiStatus := status{}

	err = json.Unmarshal(js, &apiStatus)
	if err != nil {
		return err
	}

	if apiStatus.Status.APIRuleStatus.Code == "ERROR" {
		return retry.Do(func() error {
			res, err := onRetry(k8sClient, resources...)
			if err != nil {
				return err
			}

			js, err := json.Marshal(res)
			if err != nil {
				return err
			}
			err = json.Unmarshal(js, &apiStatus)
			if err != nil {
				return err
			}
			if apiStatus.Status.APIRuleStatus.Code == "ERROR" {
				return errors.New("APIRule status not ok")
			}
			return nil
		}, h.opts...)
	}
	return nil
}
