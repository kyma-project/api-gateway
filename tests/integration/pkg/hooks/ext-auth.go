package hooks

import (
	_ "embed"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"k8s.io/client-go/dynamic"
)

//go:embed manifests/ext-auth.yaml
var extAuthYaml []byte

func applyExtAuthorizer(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	resources, err := manifestprocessor.ParseYaml(extAuthYaml)
	if err != nil {
		return err
	}

	_, err = resourceMgr.CreateOrUpdateResources(k8sClient, resources...)
	if err != nil {
		return err
	}
	return nil
}

func ApplyExtAuthorizerHook(t testcontext.Testsuite) func() error {
	return func() error {
		return applyExtAuthorizer(t.ResourceManager(), t.K8sClient())
	}
}
