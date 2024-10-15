package helpers

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type apiRuleStatus struct {
	Status struct {
		APIRuleStatus struct {
			Code        string `json:"code"`
			Description string `json:"desc"`
		} `json:"APIRuleStatus"`
	} `json:"status"`
}

type apiRuleStatusV2Alpha1 struct {
	Status struct {
		State       string `json:"state"`
		Description string `json:"description"`
	} `json:"status"`
}

// RetryableApiRule wraps any function that modifies or creates an APIRule
type RetryableApiRule func(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error)

// APIRuleWithRetries tries toExecute function and retries with onRetry if APIRule status is "ERROR" or "Error"
func ApplyApiRule(toExecute RetryableApiRule, onRetry RetryableApiRule, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured) error {
	res, err := toExecute(k8sClient, resources...)
	if err != nil {
		return err
	}

	var arVersion, code, description string

	if res.Object != nil && res.Object["apiVersion"] != nil {
		apiVersion := strings.Split(res.Object["apiVersion"].(string), "/")

		if apiVersion[1] == "v1beta1" {
			arVersion = apiVersion[1]
			arStatus, err := GetAPIRuleStatus(res)
			if err != nil {
				return err
			}
			code = arStatus.Status.APIRuleStatus.Code
		} else if apiVersion[1] == "v2alpha1" {
			arVersion = apiVersion[1]
			arStatus, err := GetAPIRuleStatusV2Alpha1(res)
			if err != nil {
				return err
			}
			code = arStatus.Status.State
		}
	}

	if code == "ERROR" || code == "Error" {
		return retry.Do(func() error {
			res, err := onRetry(k8sClient, resources...)
			if err != nil {
				return err
			}
			js, err := json.Marshal(res)
			if err != nil {
				return err
			}
			if arVersion == "v1beta1" {
				var arStatus apiRuleStatus
				err = json.Unmarshal(js, &arStatus)
				if err != nil {
					return err
				}
				code = arStatus.Status.APIRuleStatus.Code
				description = arStatus.Status.APIRuleStatus.Description
			} else if arVersion == "v2alpha1" {
				var arStatus apiRuleStatusV2Alpha1
				err = json.Unmarshal(js, &arStatus)
				if err != nil {
					return err
				}
				code = arStatus.Status.State
				description = arStatus.Status.Description
			}
			if code == "ERROR" || code == "Error" {
				log.Println("APIRule status not ok: " + description)
				return errors.New("APIRule status not ok: " + description)
			}
			return nil
		}, retryOpts...)
	}
	return nil
}

func ApplyApiRuleV2Alpha1(toExecute RetryableApiRule, onRetry RetryableApiRule, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured) error {
	res, err := toExecute(k8sClient, resources...)
	if err != nil {
		return err
	}
	apiStatus, err := GetAPIRuleStatusV2Alpha1(res)
	if err != nil {
		return err
	}
	if apiStatus.Status.State != "Ready" {
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
			if apiStatus.Status.State != "Ready" {
				log.Println("APIRule status not Ready: " + apiStatus.Status.Description)
				return errors.New("APIRule status not Ready: " + apiStatus.Status.Description)
			}
			return nil
		}, retryOpts...)
	}
	return nil
}

func ApplyApiRuleV2Alpha1ExpectError(toExecute RetryableApiRule, onRetry RetryableApiRule, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured, errorMessage string) error {
	res, err := toExecute(k8sClient, resources...)
	if err != nil {
		return err
	}
	apiStatus, err := GetAPIRuleStatusV2Alpha1(res)
	if err != nil {
		return err
	}
	if apiStatus.Status.State != "Error" {
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
			if apiStatus.Status.State != "Error" {
				log.Println("APIRule status not Error: " + apiStatus.Status.Description)
				return errors.New("APIRule status not Error: " + apiStatus.Status.Description)
			}
			if !strings.Contains(apiStatus.Status.Description, errorMessage) {
				log.Println("APIRule Error status description does not contain expected string: " + apiStatus.Status.Description)
				return errors.New("APIRule Error status description does not contain expected string: " + apiStatus.Status.Description)
			}
			return nil
		}, retryOpts...)
	}
	return nil
}

func UpdateApiRule(resourceManager *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured) error {
	status := apiRuleStatus{}

	res, err := resourceManager.UpdateResources(k8sClient, resources...)
	if err != nil {
		return err
	}

	js, err := json.Marshal(res)
	if err != nil {
		return err
	}
	err = json.Unmarshal(js, &status)
	if err != nil {
		return err
	}
	if status.Status.APIRuleStatus.Code == "ERROR" || status.Status.APIRuleStatus.Code == "Error" {
		log.Println("APIRule status not ok: " + status.Status.APIRuleStatus.Description)
		return errors.New("APIRule status not ok: " + status.Status.APIRuleStatus.Description)
	}
	return nil
}

func GetAPIRuleStatus(apiRuleUnstructured *unstructured.Unstructured) (apiRuleStatus, error) {
	js, err := json.Marshal(apiRuleUnstructured)
	if err != nil {
		return apiRuleStatus{}, err
	}

	status := apiRuleStatus{}

	err = json.Unmarshal(js, &status)
	if err != nil {
		return apiRuleStatus{}, err
	}

	return status, nil
}

func GetAPIRuleStatusV2Alpha1(apiRuleUnstructured *unstructured.Unstructured) (apiRuleStatusV2Alpha1, error) {
	js, err := json.Marshal(apiRuleUnstructured)
	if err != nil {
		return apiRuleStatusV2Alpha1{}, err
	}

	status := apiRuleStatusV2Alpha1{}

	err = json.Unmarshal(js, &status)
	if err != nil {
		return apiRuleStatusV2Alpha1{}, err
	}

	return status, nil
}

func HasAPIRuleStatus(apiRuleUnstructured *unstructured.Unstructured, status string) (bool, error) {
	apiRuleStatus, err := GetAPIRuleStatus(apiRuleUnstructured)
	if err != nil {
		return false, err
	}
	return apiRuleStatus.Status.APIRuleStatus.Code == status, nil
}
