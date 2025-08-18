package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type APIRuleStatus interface {
	GetStatus() string
	GetDescription() string
}
type APIRuleStatusV1beta1 struct {
	Status struct {
		APIRuleStatus struct {
			Code        string `json:"code"`
			Description string `json:"desc"`
		} `json:"APIRuleStatus"`
	} `json:"status"`
}

type APIRuleStatusV2alpha1 struct {
	Status struct {
		State       string `json:"state"`
		Description string `json:"description"`
	} `json:"status"`
}
type APIRuleStatusV2 struct {
	Status struct {
		State       string `json:"state"`
		Description string `json:"description"`
	} `json:"status"`
}

func (ar APIRuleStatusV1beta1) GetStatus() string {
	return ar.Status.APIRuleStatus.Code
}

func (ar APIRuleStatusV1beta1) GetDescription() string {
	return ar.Status.APIRuleStatus.Description
}

func (ar APIRuleStatusV2) GetStatus() string {
	return ar.Status.State
}

func (ar APIRuleStatusV2) GetDescription() string {
	return ar.Status.Description
}

func (ar APIRuleStatusV2alpha1) GetStatus() string {
	return ar.Status.State
}

func (ar APIRuleStatusV2alpha1) GetDescription() string {
	return ar.Status.Description
}

func GetAPIRuleStatus(res *unstructured.Unstructured) (APIRuleStatus, error) {
	apiRuleName := res.GetName()
	if res.Object == nil || res.Object["apiVersion"] == nil {
		return nil, fmt.Errorf("apiVersion not found in the APIRule %s object", apiRuleName)
	}

	apiVersion := strings.Split(res.Object["apiVersion"].(string), "/")
	apiRuleVersionVersion := apiVersion[1]

	switch apiRuleVersionVersion {
	case "v1beta1":
		arStatus, err := getAPIRuleStatusV1beta1(res)
		if err != nil {
			return nil, err
		}
		return arStatus, nil
	case "v2alpha1":
		arStatus, err := getAPIRuleStatusV2Alpha1(res)
		if err != nil {
			return nil, err
		}
		return arStatus, nil
	case "v2":
		arStatus, err := getAPIRuleStatusV2(res)
		if err != nil {
			return nil, err
		}
		return arStatus, nil
	default:
		return nil, fmt.Errorf("APIRule %s has unsupported version", apiRuleName)
	}
}

// CreateApiRule creates APIRule and waits for its status
// This function works for both v1beta1 and v2alpha1 versions of APIRule.
func CreateApiRule(resourceMgr *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, apiRuleResource unstructured.Unstructured) error {
	if apiRuleResource.GetObjectKind().GroupVersionKind().Kind != "APIRule" {
		return fmt.Errorf("object with name %s is not an APIRule unintended usage of the function", apiRuleResource.GetName())
	}

	resourceSchema, ns, _ := resourceMgr.GetResourceSchemaAndNamespace(apiRuleResource)
	apiRuleName := apiRuleResource.GetName()

	err := retry.Do(func() error {
		err := resourceMgr.CreateResource(k8sClient, resourceSchema, ns, apiRuleResource)
		if err != nil {
			return err
		}
		return nil
	}, retryOpts...)
	if err != nil {
		return fmt.Errorf("failed to create APIRule %s: %w", apiRuleName, err)
	}

	return retry.Do(func() error {
		currentApiRule, err := resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
		if err != nil {
			return err
		}

		st, err := GetAPIRuleStatus(currentApiRule)
		if err != nil {
			return err
		}

		if st.GetStatus() == "" {
			return fmt.Errorf("APIRule %s not reconciled", apiRuleName)
		}

		if strings.ToLower(st.GetStatus()) == "error" {
			log.Printf("APIRule %s status code is '%s' with description '%s'", apiRuleName, st.GetStatus(), st.GetDescription())
			return fmt.Errorf("APIRule %s in error state", apiRuleName)
		}
		return nil
	}, retryOpts...)
}

func CreateApiRuleExpectError(resourceMgr *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, apiRuleResource unstructured.Unstructured, errorMessage string) error {
	if apiRuleResource.GetObjectKind().GroupVersionKind().Kind != "APIRule" {
		return errors.New("object is not an APIRule unintended usage of the function")
	}

	resourceSchema, ns, _ := resourceMgr.GetResourceSchemaAndNamespace(apiRuleResource)
	apiRuleName := apiRuleResource.GetName()

	err := retry.Do(func() error {
		err := resourceMgr.CreateResource(k8sClient, resourceSchema, ns, apiRuleResource)
		if err != nil {
			return err
		}
		return nil
	}, retryOpts...)
	if err != nil {
		return fmt.Errorf("failed to create APIRule %s: %w", apiRuleName, err)
	}

	return retry.Do(func() error {
		currentApiRule, err := resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
		if err != nil {
			return err
		}
		apiStatus, err := GetAPIRuleStatus(currentApiRule)
		if err != nil {
			return err
		}
		if strings.ToLower(apiStatus.GetStatus()) != "error" {
			log.Printf("expected APIRule %s status to be Error, got %s with desc: %s", apiRuleName, apiStatus.GetStatus(), apiStatus.GetDescription())
			return fmt.Errorf("expected APIRule %s status to be Error, got %s with desc: %s", apiRuleName, apiStatus.GetStatus(), apiStatus.GetDescription())
		}
		if !strings.Contains(apiStatus.GetDescription(), errorMessage) {
			log.Printf("expected error description of the APIRule %s to be %s, got %s", apiRuleName, errorMessage, apiStatus.GetDescription())
			return fmt.Errorf("expected error description of the APIRule %s to be %s, got %s", apiRuleName, errorMessage, apiStatus.GetDescription())
		}
		return nil
	}, retryOpts...)
}

func UpdateApiRule(resourceMgr *resource.Manager, k8sClient dynamic.Interface, retryOpts []retry.Option, apiRuleResource unstructured.Unstructured) error {
	resourceSchema, ns, _ := resourceMgr.GetResourceSchemaAndNamespace(apiRuleResource)
	apiRuleName := apiRuleResource.GetName()

	err := resourceMgr.UpdateResource(k8sClient, resourceSchema, ns, apiRuleName, apiRuleResource)
	if err != nil {
		return err
	}

	currentApiRule, err := resourceMgr.GetResource(k8sClient, resourceSchema, ns, apiRuleName)
	if err != nil {
		return err
	}

	st, err := GetAPIRuleStatus(currentApiRule)
	if err != nil {
		return err
	}
	state := strings.ToLower(st.GetStatus())

	if !slices.Contains([]string{"ready", "ok", "warning"}, state) {
		log.Printf("APIRule status not ok, ready or warning: %s, %s\n", st.GetStatus(), st.GetDescription())
		return fmt.Errorf("status not ok, ready or warning: %s, %s", st.GetStatus(), st.GetDescription())
	}
	return nil
}

func getAPIRuleStatusV1beta1(apiRuleUnstructured *unstructured.Unstructured) (APIRuleStatusV1beta1, error) {
	js, err := json.Marshal(apiRuleUnstructured)
	if err != nil {
		return APIRuleStatusV1beta1{}, err
	}

	status := APIRuleStatusV1beta1{}

	err = json.Unmarshal(js, &status)
	if err != nil {
		return APIRuleStatusV1beta1{}, err
	}

	return status, nil
}

func getAPIRuleStatusV2Alpha1(apiRuleUnstructured *unstructured.Unstructured) (APIRuleStatusV2alpha1, error) {
	js, err := json.Marshal(apiRuleUnstructured)
	if err != nil {
		return APIRuleStatusV2alpha1{}, err
	}

	status := APIRuleStatusV2alpha1{}

	err = json.Unmarshal(js, &status)
	if err != nil {
		return APIRuleStatusV2alpha1{}, err
	}

	return status, nil
}

func getAPIRuleStatusV2(apiRuleUnstructured *unstructured.Unstructured) (APIRuleStatusV2, error) {
	js, err := json.Marshal(apiRuleUnstructured)
	if err != nil {
		return APIRuleStatusV2{}, err
	}

	status := APIRuleStatusV2{}

	err = json.Unmarshal(js, &status)
	if err != nil {
		return APIRuleStatusV2{}, err
	}

	return status, nil
}
