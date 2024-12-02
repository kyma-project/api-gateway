package global

import (
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"log"
)

const globalCommonsFileName = "global-commons.yaml"

func GenerateNamespaceName(testsuiteName string) string {
	return fmt.Sprintf("%s-%s", testsuiteName, helpers.GenerateRandomString())
}

func getGlobalResources(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, manifestsDir string) ([]unstructured.Unstructured, error) {
	return manifestprocessor.ParseFromFileWithTemplate(globalCommonsFileName, manifestsDir, struct {
		Namespace string
	}{
		Namespace: namespace,
	})
}

func CreateGlobalResources(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, manifestsDir string) error {
	resources, err := getGlobalResources(resourceMgr, k8sClient, namespace, manifestsDir)
	if err != nil {
		return fmt.Errorf("error getting common tests resources: %w", err)
	}

	log.Printf("Creating common tests resources")
	_, err = resourceMgr.CreateResources(k8sClient, resources...)
	if err != nil {
		return err
	}
	log.Printf("Common tests resources created")
	return nil
}

func DeleteGlobalResources(resourceMgr *resource.Manager, k8sClient dynamic.Interface, namespace string, manifestsDir string) error {
	resources, err := getGlobalResources(resourceMgr, k8sClient, namespace, manifestsDir)
	if err != nil {
		return fmt.Errorf("error getting common tests resources: %w", err)
	}

	log.Printf("Deleting common tests resources")
	err = resourceMgr.DeleteResources(k8sClient, resources...)
	if err != nil {
		return err
	}
	log.Printf("Common tests resources deleted")
	return nil
}
