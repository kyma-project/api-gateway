package hooks

import (
	_ "embed"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/istio"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

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
