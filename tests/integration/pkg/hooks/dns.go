package hooks

import (
	"github.com/kyma-project/api-gateway/tests/integration/pkg/network"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

func DnsPatchForK3dSuiteHook(t testcontext.Testsuite) func() error {
	return func() error {
		return network.CreateKymaLocalDnsRewrite(t.ResourceManager(), t.K8sClient())
	}
}

func DnsPatchForK3dSuiteHookTeardown(t testcontext.Testsuite) func() error {
	return func() error {
		return network.DeleteKymaLocalDnsRewrite(t.ResourceManager(), t.K8sClient())
	}
}
