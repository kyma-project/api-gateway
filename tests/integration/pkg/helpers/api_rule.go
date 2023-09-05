package helpers

import (
	"encoding/json"
	"errors"
	"github.com/avast/retry-go/v4"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// RetryableApiRule wraps any function that modifies or creates an APIRule
type RetryableApiRule func(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error)

// APIRuleWithRetries tries toExecute function and retries with onRetry if APIRule status is "ERROR"
func ApplyApiRule(toExecute RetryableApiRule, onRetry RetryableApiRule, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured) error {
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
		}, retryOpts...)
	}
	return nil
}
