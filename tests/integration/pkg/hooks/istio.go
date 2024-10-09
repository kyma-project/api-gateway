package hooks

import (
	_ "embed"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/istio"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

//go:embed manifests/ext-auth.yaml
var extAuthYaml []byte

const skipVerifyJwksResolverEnvVar = "JWKS_RESOLVER_INSECURE_SKIP_VERIFY"
const extAuthorizerName = "sample-ext-authz-http"
const extAuthorizerPort = 8000
const extAuthorizerService = "ext-authz.ext-auth.svc.cluster.local"
const extAuthorizerHeader = "x-ext-authz"

func IstioSkipVerifyJwksResolverSuiteHook(t testcontext.Testsuite) func() error {
	return func() error {
		envVar := map[string]string{skipVerifyJwksResolverEnvVar: "true"}
		return istio.PatchIstiodDeploymentWithEnvironmentVariables(t.ResourceManager(), t.K8sClient(), envVar)
	}
}

func IstioSkipVerifyJwksResolverSuiteHookTeardown(t testcontext.Testsuite) func() error {
	return func() error {
		return istio.RemoveEnvironmentVariableFromIstiodDeployment(t.ResourceManager(), t.K8sClient(), skipVerifyJwksResolverEnvVar)
	}
}

func installExternalAuthorizer(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	log.Printf("Patching Istio CR with external authorizer")
	res, err := resourceMgr.GetResource(k8sClient, schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha2", Resource: "istios"}, "kyma-system", "default", []retry.Option{retry.Attempts(1)}...)
	if err != nil {
		return fmt.Errorf("Could not get Istio CR: %s", err.Error())
	}

	authorizers, _, err := unstructured.NestedSlice(res.Object, "spec", "config", "authorizers")
	if err != nil {
		return fmt.Errorf("Could not get authorizers: %s", err.Error())
	}

	var newAuthorizer = map[string]interface{}{
		"name":    extAuthorizerName,
		"port":    int64(extAuthorizerPort),
		"service": extAuthorizerService,
		"headers": map[string]interface{}{
			"inCheck": map[string]interface{}{
				"include": append(make([]interface{}, 0), extAuthorizerHeader),
			},
		},
	}

	authorizerFound := false
	for i, authorizer := range authorizers {
		authorizerName, _, err := unstructured.NestedString(authorizer.(map[string]interface{}), "name")
		if err != nil {
			return fmt.Errorf("Could get authorizer name: %s", err.Error())
		}
		if authorizerName == extAuthorizerName {
			log.Printf("Authorizer already exists, updating it to an expected state")
			authorizers[i] = newAuthorizer
			authorizerFound = true
		}
	}
	if !authorizerFound {
		authorizers = append(authorizers, newAuthorizer)
	}

	err = unstructured.SetNestedSlice(res.Object, authorizers, "spec", "config", "authorizers")
	if err != nil {
		return fmt.Errorf("could not set authorizers: %s", err.Error())
	}

	err = resourceMgr.UpdateResource(k8sClient, schema.GroupVersionResource{Group: "operator.kyma-project.io", Version: "v1alpha2", Resource: "istios"}, "kyma-system", "default", *res)
	if err != nil {
		return fmt.Errorf("could not update istio CR: %s", err.Error())
	}

	log.Printf("Deploying External Authorizer")
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

func ExternalAuthorizerInstallHook(t testcontext.Testsuite) func() error {
	return func() error {
		return installExternalAuthorizer(t.ResourceManager(), t.K8sClient())
	}
}
