package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type ApiRuleStatusV1beta1 struct {
	Status struct {
		APIRuleStatus struct {
			Code        string `json:"code"`
			Description string `json:"desc"`
		} `json:"APIRuleStatus"`
	} `json:"status"`
}

type ApiRuleStatusV2alpha1 struct {
	Status struct {
		State       string `json:"state"`
		Description string `json:"description"`
	} `json:"status"`
}

const (
	errorV1beta1      = "ERROR"
	errorV2alpha1     = "Error"
	notReconciledCode = ""
)

// RetryableApiRule wraps any function that modifies or creates an APIRule
type RetryableApiRule func(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error)

func getAPIRuleStatus(res *unstructured.Unstructured) (string, string, error) {
	if res.Object == nil || res.Object["apiVersion"] == nil {
		return "", "", errors.New("apiVersion not found in the APIRule object")
	}

	apiVersion := strings.Split(res.Object["apiVersion"].(string), "/")
	apiRuleVersionVersion := apiVersion[1]

	var description, code string

	if apiRuleVersionVersion == "v1beta1" {
		arStatus, err := GetAPIRuleStatusV1beta1(res)
		if err != nil {
			return "", "", err
		}
		code = arStatus.Status.APIRuleStatus.Code
		description = arStatus.Status.APIRuleStatus.Description
	} else if apiRuleVersionVersion == "v2alpha1" {
		arStatus, err := GetAPIRuleStatusV2Alpha1(res)
		if err != nil {
			return "", "", err
		}
		code = arStatus.Status.State
		description = arStatus.Status.Description
	}

	return code, description, nil
}

// ApplyApiRule tries toExecute function and retries with onRetry if APIRule status is in error status or has no status
// code. This function works for both v1beta1 and v2alpha1 versions of APIRule.
func ApplyApiRule(toExecute RetryableApiRule, onRetry RetryableApiRule, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured) error {
	res, err := toExecute(k8sClient, resources...)
	if err != nil {
		return err
	}
	code, _, err := getAPIRuleStatus(res)
	if err != nil {
		return err
	}

	if code == errorV1beta1 || code == errorV2alpha1 || code == notReconciledCode {
		return retry.Do(func() error {
			res, err := onRetry(k8sClient, resources...)
			if err != nil {
				return err
			}

			code, description, err := getAPIRuleStatus(res)
			if err != nil {
				return err
			}

			switch code {
			case notReconciledCode:
				return errors.New("apirule not reconciled")
			case errorV1beta1, errorV2alpha1:
				log.Printf("APIRule status code is '%s' with description '%s'", code, description)
				return errors.New("apirule in error state")
			default:
				return nil
			}
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
				log.Printf("expected but APIRule status to be Error, got %s with desc: %s", apiStatus.Status.State, apiStatus.Status.Description)
				return fmt.Errorf("expected but APIRule status to be Error, got %s with desc: %s", apiStatus.Status.State, apiStatus.Status.Description)
			}
			if !strings.Contains(apiStatus.Status.Description, errorMessage) {
				log.Printf("expected error description of the APIRule to be %s, got %s", errorMessage, apiStatus.Status.Description)
				return fmt.Errorf("expected error description of the APIRule to be %s, got %s", errorMessage, apiStatus.Status.Description)
			}
			return nil
		}, retryOpts...)
	}
	return nil
}

func UpdateApiRule(resourceManager *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured) error {
	status := ApiRuleStatusV1beta1{}

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

func GetAPIRuleStatusV1beta1(apiRuleUnstructured *unstructured.Unstructured) (ApiRuleStatusV1beta1, error) {
	js, err := json.Marshal(apiRuleUnstructured)
	if err != nil {
		return ApiRuleStatusV1beta1{}, err
	}

	status := ApiRuleStatusV1beta1{}

	err = json.Unmarshal(js, &status)
	if err != nil {
		return ApiRuleStatusV1beta1{}, err
	}

	return status, nil
}

func GetAPIRuleStatusV2Alpha1(apiRuleUnstructured *unstructured.Unstructured) (ApiRuleStatusV2alpha1, error) {
	js, err := json.Marshal(apiRuleUnstructured)
	if err != nil {
		return ApiRuleStatusV2alpha1{}, err
	}

	status := ApiRuleStatusV2alpha1{}

	err = json.Unmarshal(js, &status)
	if err != nil {
		return ApiRuleStatusV2alpha1{}, err
	}

	return status, nil
}

func HasAPIRuleStatus(apiRuleUnstructured *unstructured.Unstructured, status string) (bool, error) {
	apiRuleStatus, err := GetAPIRuleStatusV1beta1(apiRuleUnstructured)
	if err != nil {
		return false, err
	}
	return apiRuleStatus.Status.APIRuleStatus.Code == status, nil
}
