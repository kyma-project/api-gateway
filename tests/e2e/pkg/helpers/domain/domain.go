package domain

import (
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"testing"
)

func GetFromGateway(t *testing.T, gatewayName, gatewayNamespace string) (string, error) {
	t.Helper()
	r, err := infrastructure.ResourcesClient(t)
	if err != nil {
		t.Logf("Failed to get resources client: %v", err)
		return "", err
	}

	gateway := &v1alpha3.Gateway{}
	err = r.Get(t.Context(), gatewayName, gatewayNamespace, gateway)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			t.Logf("Gateway %s not found in namespace %s", gatewayName, gatewayNamespace)
			return "", nil
		}
		t.Logf("Failed to get Gateway %s in namespace %s: %v", gatewayName, gatewayNamespace, err)
		return "", err
	}

	servers := gateway.Spec.GetServers()
	if len(servers) == 0 {
		t.Logf("No servers found in Gateway %s in namespace %s", gatewayName, gatewayNamespace)
		return "", nil
	}

	if len(servers) > 1 {
		t.Logf("Multiple servers found in Gateway %s in namespace %s, returning the first one", gatewayName, gatewayNamespace)
	}

	hosts := servers[0].GetHosts()
	if len(hosts) == 0 {
		t.Logf("No hosts found in the first server of Gateway %s in namespace %s", gatewayName, gatewayNamespace)
		return "", nil
	}

	if len(hosts) > 1 {
		t.Logf("Multiple hosts found in the first server of Gateway %s in namespace %s, returning the first one", gatewayName, gatewayNamespace)
	}

	wildcardHost := hosts[0]
	t.Logf("Wildcard host from Gateway %s in namespace %s: %s", gatewayName, gatewayNamespace, wildcardHost)

	return strings.TrimPrefix(wildcardHost, "*."), nil
}
