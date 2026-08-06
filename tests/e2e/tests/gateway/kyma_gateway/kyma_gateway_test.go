package kyma_gateway

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "embed"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/resourceasserts"
	e2eclient "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	httpbinhelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpbin"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/v1access"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

const checkTimeout = 2 * time.Minute

//go:embed vs.yaml
var VirtualService string

type resourceCheck struct {
	gvk       schema.GroupVersionKind
	name      string
	namespace string
}

var oathkeeperCRDResource = resourceCheck{gvk: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, name: "rules.oathkeeper.ory.sh", namespace: ""}

var oathkeeperResources = []resourceCheck{
	{gvk: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, name: "ory-oathkeeper", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, name: "ory-oathkeeper-config", namespace: "kyma-system"},
	oathkeeperCRDResource,
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}, name: "ory-oathkeeper-jwks-secret", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, name: "ory-oathkeeper-api", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, name: "ory-oathkeeper-proxy", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, name: "ory-oathkeeper-maester-metrics", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}, name: "ory-oathkeeper", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}, name: "oathkeeper-maester-account", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"}, name: "oathkeeper-maester-role", namespace: ""},
	{gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"}, name: "oathkeeper-maester-role-binding", namespace: ""},
	{gvk: schema.GroupVersionKind{Group: "security.istio.io", Version: "v1beta1", Kind: "PeerAuthentication"}, name: "ory-oathkeeper-maester-metrics", namespace: "kyma-system"},
	{gvk: schema.GroupVersionKind{Group: "policy", Version: "v1", Kind: "PodDisruptionBudget"}, name: "ory-oathkeeper", namespace: "kyma-system"},
}

func TestKymaGateway(t *testing.T) {
	t.Run("Oathkeeper is installed and uninstalled depending on APIGateway presence", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r, t))
		modulehelpers.SetupBaseCR(t)

		kymaGatewayResource := resourceCheck{gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"}, name: "kyma-gateway", namespace: "kyma-system"}
		resourceasserts.AssertResourceExists(t, r, resourceasserts.StructCheck{Gvk: kymaGatewayResource.gvk, Name: kymaGatewayResource.name, Namespace: kymaGatewayResource.namespace}, checkTimeout, time.Second)
		require.NoError(t, wait.For(
			conditions.New(r).DeploymentAvailable("ory-oathkeeper", "kyma-system"),
			wait.WithTimeout(2*time.Minute),
		), "Deployment %s/%s should be Ready", "kyma-system", "ory-oathkeeper")

		for _, rc := range oathkeeperResources {
			resourceasserts.AssertResourceExists(t, r, resourceasserts.StructCheck{Gvk: rc.gvk, Name: rc.name, Namespace: rc.namespace}, checkTimeout, time.Second)
		}

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, "kyma-system", "default"))
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayDeleted(t, r, "kyma-system", "default"))

		for _, rc := range oathkeeperResources {
			resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: rc.gvk, Name: rc.name, Namespace: rc.namespace}, checkTimeout, time.Second)
		}
	})

	t.Run("Oathkeeper is not installed when there is no api-gateway-config ConfigMap", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r, t))
		modulehelpers.SetupBaseCR(t)

		cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: v1access.V1AccessConfigMapName, Namespace: v1access.V1AccessConfigMapNamespace}}
		require.NoError(t, r.Delete(context.Background(), &cm))

		accessCm := resourceCheck{gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, name: v1access.V1AccessConfigMapName, namespace: v1access.V1AccessConfigMapNamespace}
		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: accessCm.gvk, Name: accessCm.name, Namespace: accessCm.namespace}, checkTimeout, time.Second)
		// The integration feature describes APIGateway removal first. In practice, without waiting for
		// asynchronous deletion to complete, the effective order is nondeterministic. This test uses a
		// deterministic sequence: remove the access ConfigMap first, then delete APIGateway with an
		// explicit wait, so the expected Oathkeeper Rule CRD state can be asserted reliably.
		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, "kyma-system", "default"))
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayDeleted(t, r, "kyma-system", "default"))

		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
		for _, rc := range oathkeeperResources {
			if rc.gvk.Kind == oathkeeperCRDResource.gvk.Kind && rc.name == oathkeeperCRDResource.name {
				continue
			}
			resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: rc.gvk, Name: rc.name, Namespace: rc.namespace}, checkTimeout, time.Second)
		}
		resourceasserts.AssertResourceExists(t, r, resourceasserts.StructCheck{Gvk: oathkeeperCRDResource.gvk, Name: oathkeeperCRDResource.name, Namespace: oathkeeperCRDResource.namespace}, checkTimeout, time.Second)
	})

	t.Run("Kyma Gateway is not removed when there is a VirtualService", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		svcName, _, err := httpbinhelpers.DeployHttpbin(t, "kyma-system")
		require.NoError(t, err, "Failed to deploy httpbin service")

		vsName := "kyma-vs"
		vs, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			VirtualService,
			map[string]any{"Name": vsName, "Namespace": "kyma-system", "Host": "local.kyma.dev", "Gateway": fmt.Sprintf("kyma-system/kyma-gateway"), "DestinationHost": fmt.Sprintf("%s.kyma-system.svc.cluster.local", svcName)},
		)
		require.NoError(t, err, "Failed to create VirtualService resource")

		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		cr := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "kyma-system"}}
		require.NoError(t, r.Get(context.Background(), cr.GetName(), cr.GetNamespace(), &cr), "Failed to get APIGateway CR")
		cr.Spec.EnableKymaGateway = new(false)
		require.NoError(t, r.Update(context.Background(), &cr), "Failed to update APIGateway CR to disable Kyma Gateway")

		istioGW := &unstructured.Unstructured{}
		istioGW.SetGroupVersionKind(schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"})
		istioGW.SetName("kyma-gateway")
		istioGW.SetNamespace("kyma-system")
		require.NoError(t, r.Get(context.Background(), "kyma-gateway", "kyma-system", istioGW))

		require.NoError(t, modulehelpers.WaitUntilAPIGatewayExists(t, r, cr.GetNamespace(), cr.GetName()), "APIGateway CR should still exist while VirtualService is present")
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayCRHasState(t, r, cr.GetNamespace(), cr.GetName(), operatorv1alpha1.Warning, "There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"), "APIGateway CR should be in Warning state while VirtualService is present")

		require.NoError(t, r.Get(context.Background(), "kyma-gateway", "kyma-system", istioGW))
		require.NoError(t, r.Delete(context.Background(), vs), "Failed to delete VirtualService resource")

		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "VirtualService"}, Name: vs.GetName(), Namespace: vs.GetNamespace()}, checkTimeout, time.Second)
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayCRHasState(t, r, cr.GetNamespace(), cr.GetName(), operatorv1alpha1.Ready, "Successfully reconciled"), "APIGateway CR should be in Ready state after VirtualService is removed")
	})

	t.Run("Kyma Gateway is removed when there is no blocking resources", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, "kyma-system", "default"))
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayDeleted(t, r, "kyma-system", "default"))

		gw := resourceCheck{gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"}, name: "kyma-gateway", namespace: "kyma-system"}
		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: gw.gvk, Name: gw.name, Namespace: gw.namespace}, checkTimeout, time.Second)
	})

	t.Run("Second APIGateway CR is applied to the cluster", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		secondCR := &operatorv1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{Name: "second-api-gateway-cr", Namespace: "kyma-system"},
		}
		require.NoError(t, r.Create(t.Context(), secondCR), "Failed to create second APIGateway CR")

		require.NoError(t, modulehelpers.WaitUntilAPIGatewayCRHasState(t, r, secondCR.Namespace, secondCR.Name, operatorv1alpha1.Warning, "stopped APIGateway CR reconciliation: only APIGateway CR default reconciles the module"), "Second APIGateway CR should be in Warning state")
		require.NoError(t, modulehelpers.WaitUntilAPIGatewayCRHasState(t, r, "kyma-system", "default", operatorv1alpha1.Ready, "Successfully reconciled"), "Default APIGateway CR should be in Ready state")

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, "kyma-system", "second-api-gateway-cr"), "Failed to delete second APIGateway CR")

	})
}
