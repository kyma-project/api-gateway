package hooks

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/avast/retry-go/v4"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
	"time"
)

//go:embed manifests/istio-cr.yaml
var istioCrManifest []byte

//go:embed manifests/ext-auth.yaml
var extAuthManifests []byte

func applyExtAuthorizerIstioCR() error {
	log.Printf("Apply Istio CR with External Authorizer config")
	istioCr, err := getIstioCr()
	if err != nil {
		return err
	}
	client := k8sclient.GetK8sClient()
	spec := istioCr.Object["spec"]

	return retry.Do(func() error {
		_, err := controllerutil.CreateOrPatch(context.Background(), client, &istioCr, func() error {
			istioCr.Object["spec"] = spec
			return nil
		})
		return err
	}, []retry.Option{
		retry.Delay(time.Duration(2) * time.Second),
		retry.Attempts(5),
		retry.DelayType(retry.FixedDelay),
	}...)
}

func getIstioCr() (unstructured.Unstructured, error) {
	var istioCr unstructured.Unstructured
	err := yaml.Unmarshal(istioCrManifest, &istioCr)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	return istioCr, nil
}

func deployExtAuthorizer(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	resources, err := manifestprocessor.ParseYaml(extAuthManifests)
	if err != nil {
		return err
	}

	if len(resources) == 0 || resources[0].GetKind() != "Namespace" {
		return fmt.Errorf("First resource should be a namespace")
	}

	log.Printf("Creating External Authorizer namespace and deployment")
	_, err = resourceMgr.CreateOrUpdateResources(k8sClient, resources...)
	if err != nil {
		return err
	}
	return nil
}

func removeExtAuthorizer(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	resources, err := manifestprocessor.ParseYaml(extAuthManifests)
	if err != nil {
		return err
	}

	if len(resources) == 0 || resources[0].GetKind() != "Namespace" {
		return fmt.Errorf("First resource should be a namespace")
	}

	nsResourceSchema, ns, name := resourceMgr.GetResourceSchemaAndNamespace(resources[0])
	log.Printf("Deleting External Authorizer namespace, if exists: %s\n", name)
	err = resourceMgr.DeleteResource(k8sClient, nsResourceSchema, ns, name)
	if err != nil {
		return err
	}

	return nil
}

func ExtAuthorizerInstallHook(t testcontext.Testsuite) func() error {
	return func() error {
		err := applyExtAuthorizerIstioCR()
		if err != nil {
			return err
		}

		return deployExtAuthorizer(t.ResourceManager(), t.K8sClient())
	}
}

func ExtAuthorizerRemoveHook(t testcontext.Testsuite) func() error {
	return func() error {
		return removeExtAuthorizer(t.ResourceManager(), t.K8sClient())
	}
}
