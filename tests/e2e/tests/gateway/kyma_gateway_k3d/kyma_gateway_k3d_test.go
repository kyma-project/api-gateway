package kyma_gateway_k3d

import (
	"testing"
	"time"

	"github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/resourceasserts"
	e2eclient "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	checkTimeout = 2 * time.Minute
)

var apiGatewayResources = []resourceasserts.StructCheck{
	{Gvk: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, Name: "apigateways.operator.kyma-project.io", Namespace: ""},
	{Gvk: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, Name: "apirules.gateway.kyma-project.io", Namespace: ""},
	{Gvk: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, Name: "api-gateway-controller-manager", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}, Name: "api-gateway-controller-manager", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"}, Name: "api-gateway-leader-election-role", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"}, Name: "api-gateway-leader-election-rolebinding", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"}, Name: "api-gateway-manager-role", Namespace: ""},
	{Gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"}, Name: "api-gateway-manager-rolebinding", Namespace: ""},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, Name: "api-gateway-apirule-ui.operator.kyma-project.io", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, Name: "api-gateway-ui.operator.kyma-project.io", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, Name: "api-gateway-operator-metrics", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "scheduling.k8s.io", Version: "v1", Kind: "PriorityClass"}, Name: "api-gateway-priority-class", Namespace: ""},
	{Gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"}, Name: "kyma-gateway", Namespace: "kyma-system"},
	{Gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "VirtualService"}, Name: "istio-healthz", Namespace: "istio-system"},
}

func TestKymaGatewayK3D(t *testing.T) {
	t.Run("API Gateway is completely deployed", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		for _, res := range apiGatewayResources {
			resourceasserts.AssertResourceExists(t, r, res, checkTimeout)
		}
	})
}
