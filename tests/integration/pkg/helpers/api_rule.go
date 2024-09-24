package helpers

import (
	"encoding/json"
	"errors"
	"log"

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

// APIRuleWithRetries tries toExecute function and retries with onRetry if APIRule status is "ERROR"
func ApplyApiRule(toExecute RetryableApiRule, onRetry RetryableApiRule, k8sClient dynamic.Interface, retryOpts []retry.Option, resources []unstructured.Unstructured) error {

	res, err := toExecute(k8sClient, resources...)
	if err != nil {
		return err
	}

	apiStatus, err := GetAPIRuleStatus(res)
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
				log.Println("APIRule status not ok: " + apiStatus.Status.APIRuleStatus.Description)
				return errors.New("APIRule status not ok: " + apiStatus.Status.APIRuleStatus.Description)
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
