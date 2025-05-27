package hooks

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"

	"github.com/avast/retry-go/v4"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

const authorizerDeploymentName = "ext-authz"

//go:embed manifests/ext-auth-istio-cr.yaml
var extAuthIstioCrManifest []byte

//go:embed manifests/ext-auth.yaml
var extAuthManifests []byte

var namespaceGVK = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}

func applyExtAuthorizerIstioCR() error {
	log.Printf("Apply Istio CR with External Authorizer config")
	istioCr, err := getExtAuthIstioCr()
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
	}, testcontext.GetRetryOpts()...)
}

func getExtAuthIstioCr() (unstructured.Unstructured, error) {
	var istioCr unstructured.Unstructured
	err := yaml.Unmarshal(extAuthIstioCrManifest, &istioCr)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	return istioCr, nil
}

func getExtAuthNamespace(manifests []unstructured.Unstructured) (string, error) {
	for _, manifest := range manifests {
		if manifest.GetKind() == "Namespace" {
			return manifest.GetName(), nil
		}
	}
	return "", errors.New("there is no namespace defined in the Ext Auth resource")
}

func deployExtAuthorizer(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	resources, err := manifestprocessor.ParseYaml(extAuthManifests)
	if err != nil {
		return err
	}

	nsName, err := getExtAuthNamespace(resources)
	if err != nil {
		return err
	}

	log.Printf("Deploying External Authorizer namespace %s\n", nsName)
	_, err = resourceMgr.CreateResources(k8sClient, resources...)
	if err != nil {
		return err
	}

	log.Printf("Waiting for External Authorizer deployment")
	err = helpers.WaitForDeployment(resourceMgr, k8sClient, nsName, authorizerDeploymentName, testcontext.GetRetryOpts())
	if err != nil {
		return fmt.Errorf("external Authorizer deployment can't start")
	}
	log.Printf("External Authorizer deployment finished")

	return nil
}

func removeExtAuthorizer(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	resources, err := manifestprocessor.ParseYaml(extAuthManifests)
	if err != nil {
		return err
	}

	nsName, err := getExtAuthNamespace(resources)
	if err != nil {
		return err
	}

	log.Printf("Deleting External Authorizer namespace: %s\n", nsName)
	err = resourceMgr.DeleteResource(k8sClient, namespaceGVK, "", nsName)
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
