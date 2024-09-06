package hooks

import (
	"context"
	_ "embed"
	"github.com/avast/retry-go/v4"
	k8sclient "github.com/kyma-project/api-gateway/tests/integration/pkg/client"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/istio"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"
	"time"
)

//go:embed manifests/istio-cr.yaml
var istioCrManifest []byte

const skipVerifyJwksResolverEnvVar = "JWKS_RESOLVER_INSECURE_SKIP_VERIFY"

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

var ApplyExtAuthorizerIstioCR = func() error {
	log.Printf("Apply Istio CR with External Authorizer config")
	istioCr, err := getIstioCr()
	if err != nil {
		return err
	}
	client := k8sclient.GetK8sClient()
	spec, _ := istioCr.Object["spec"]
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
