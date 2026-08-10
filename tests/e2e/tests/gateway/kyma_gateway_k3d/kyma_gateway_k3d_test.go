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

type resourceCheck struct {
	gvk       schema.GroupVersionKind
	name      string
	namespace string
}

var apiGatewayResources = []resourceCheck{
	{gvk: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, name: "apigateways.operator.kyma-project.io", namespace: ""},
	{gvk: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, name: "apirules.gateway.kyma-project.io", namespace: ""},
	{gvk: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, name: "api-gateway-controller-manager", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}, name: "api-gateway-controller-manager", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"}, name: "api-gateway-leader-election-role", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"}, name: "api-gateway-leader-election-rolebinding", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"}, name: "api-gateway-manager-role", namespace: ""},
	{gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"}, name: "api-gateway-manager-rolebinding", namespace: ""},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, name: "api-gateway-apirule-ui.operator.kyma-project.io", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, name: "api-gateway-ui.operator.kyma-project.io", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, name: "api-gateway-operator-metrics", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "scheduling.k8s.io", Version: "v1", Kind: "PriorityClass"}, name: "api-gateway-priority-class", namespace: ""},
	{gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"}, name: "kyma-gateway", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "VirtualService"}, name: "istio-healthz", namespace: "istio-system"},
}

func TestKymaGatewayK3D(t *testing.T) {
	t.Run("API Gateway is completely deployed", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		for _, res := range apiGatewayResources {
			resourceasserts.AssertResourceExists(t, r, resourceasserts.StructCheck{Gvk: res.gvk, Name: res.name, Namespace: res.namespace}, checkTimeout)
		}
	})
}
