package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"log"
	"strings"
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

func getAPIRuleStatus(res *unstructured.Unstructured) (string, string, error) {
	apiRuleName := res.GetName()
	if res.Object == nil || res.Object["apiVersion"] == nil {
		return "", "", fmt.Errorf("apiVersion not found in the APIRule %s object", apiRuleName)
	}

	apiVersion := strings.Split(res.Object["apiVersion"].(string), "/")
	apiRuleVersionVersion := apiVersion[1]

	var description, code string

	switch apiRuleVersionVersion {
	case "v1beta1":
		arStatus, err := GetAPIRuleStatusV1beta1(res)
		if err != nil {
			return "", "", err
		}
		code = arStatus.Status.APIRuleStatus.Code
		description = arStatus.Status.APIRuleStatus.Description
	case "v2alpha1":
		arStatus, err := GetAPIRuleStatusV2Alpha1(res)
		if err != nil {
			return "", "", err
		}
		code = arStatus.Status.State
		description = arStatus.Status.Description
	default:
		return "", "", fmt.Errorf("APIRule %s has unsupported version", apiRuleName)
	}

	return code, description, nil
}

// CreateApiRule creates APIRule and waits for its status
// This function works for both v1beta1 and v2alpha1 versions of APIRule.
func CreateApiRule(resourceMgr *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, apiRuleResource unstructured.Unstructured) error {
	if apiRuleResource.GetObjectKind().GroupVersionKind().Kind != "APIRule" {
		return fmt.Errorf("object with name %s is not an APIRule unintended usage of the function", apiRuleResource.GetName())
	}

	resourceSchema, ns, _ := resourceMgr.GetResourceSchemaAndNamespace(apiRuleResource)
	apiRuleName := apiRuleResource.GetName()

	err := resourceMgr.CreateResource(k8sClient, resourceSchema, ns, apiRuleResource)
	if err != nil {
		return err
	}

	currentApiRule, err := resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
	if err != nil {
		return err
	}

	code, _, err := getAPIRuleStatus(currentApiRule)
	if err != nil {
		return err
	}

	if code == errorV1beta1 || code == errorV2alpha1 || code == notReconciledCode {
		return retry.Do(func() error {
			currentApiRule, err = resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
			if err != nil {
				return err
			}

			code, description, err := getAPIRuleStatus(currentApiRule)
			if err != nil {
				return err
			}

			switch code {
			case notReconciledCode:
				return fmt.Errorf("APIRule %s not reconciled", apiRuleName)
			case errorV1beta1, errorV2alpha1:
				log.Printf("APIRule %s status code is '%s' with description '%s'", apiRuleName, code, description)
				return fmt.Errorf("APIRule %s in error state", apiRuleName)
			default:
				return nil
			}
		}, retryOpts...)
	}
	return nil
}

func CreateApiRuleV2Alpha1(resourceMgr *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, apiRuleResource unstructured.Unstructured) error {
	if apiRuleResource.GetObjectKind().GroupVersionKind().Kind != "APIRule" {
		return fmt.Errorf("object with name %s is not an APIRule unintended usage of the function", apiRuleResource.GetName())
	}

	resourceSchema, ns, _ := resourceMgr.GetResourceSchemaAndNamespace(apiRuleResource)
	apiRuleName := apiRuleResource.GetName()

	err := resourceMgr.CreateResource(k8sClient, resourceSchema, ns, apiRuleResource)
	if err != nil {
		return err
	}

	currentApiRule, err := resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
	if err != nil {
		return err
	}
	apiStatus, err := GetAPIRuleStatusV2Alpha1(currentApiRule)
	if err != nil {
		return err
	}
	if apiStatus.Status.State != "Ready" {
		return retry.Do(func() error {
			currentApiRule, err = resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
			if err != nil {
				return err
			}

			js, err := json.Marshal(currentApiRule)
			if err != nil {
				return err
			}
			err = json.Unmarshal(js, &apiStatus)
			if err != nil {
				return err
			}
			if apiStatus.Status.State != "Ready" {
				log.Printf("APIRule %s status not Ready, but is: %s\n", apiRuleName, apiStatus.Status.Description)
				return fmt.Errorf("APIRule %s status not Ready, but is: %s", apiRuleName, apiStatus.Status.Description)
			}
			return nil
		}, retryOpts...)
	}
	return nil
}

func CreateApiRuleV2Alpha1ExpectError(resourceMgr *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, apiRuleResource unstructured.Unstructured, errorMessage string) error {
	if apiRuleResource.GetObjectKind().GroupVersionKind().Kind != "APIRule" {
		return errors.New("object is not an APIRule unintended usage of the function")
	}

	resourceSchema, ns, _ := resourceMgr.GetResourceSchemaAndNamespace(apiRuleResource)
	apiRuleName := apiRuleResource.GetName()

	err := resourceMgr.CreateResource(k8sClient, resourceSchema, ns, apiRuleResource)
	if err != nil {
		return err
	}

	currentApiRule, err := resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
	if err != nil {
		return err
	}

	apiStatus, err := GetAPIRuleStatusV2Alpha1(currentApiRule)
	if err != nil {
		return err
	}
	if apiStatus.Status.State != "Error" {
		return retry.Do(func() error {
			currentApiRule, err = resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
			if err != nil {
				return err
			}
			js, err := json.Marshal(currentApiRule)
			if err != nil {
				return err
			}
			err = json.Unmarshal(js, &apiStatus)
			if err != nil {
				return err
			}
			if apiStatus.Status.State != "Error" {
				log.Printf("expected APIRule %s status to be Error, got %s with desc: %s", apiRuleName, apiStatus.Status.State, apiStatus.Status.Description)
				return fmt.Errorf("expected APIRule %s status to be Error, got %s with desc: %s", apiRuleName, apiStatus.Status.State, apiStatus.Status.Description)
			}
			if !strings.Contains(apiStatus.Status.Description, errorMessage) {
				log.Printf("expected error description of the APIRule %s to be %s, got %s", apiRuleName, errorMessage, apiStatus.Status.Description)
				return fmt.Errorf("expected error description of the APIRule %s to be %s, got %s", apiRuleName, errorMessage, apiStatus.Status.Description)
			}
			return nil
		}, retryOpts...)
	}
	return nil
}

func UpdateApiRule(resourceMgr *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, apiRuleResource unstructured.Unstructured) error {
	resourceSchema, ns, _ := resourceMgr.GetResourceSchemaAndNamespace(apiRuleResource)
	apiRuleName := apiRuleResource.GetName()

	status := ApiRuleStatusV1beta1{}

	err := resourceMgr.UpdateResource(k8sClient, resourceSchema, ns, apiRuleName, apiRuleResource)
	if err != nil {
		return err
	}

	currentApiRule, err := resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
	if err != nil {
		return err
	}

	js, err := json.Marshal(currentApiRule)
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
