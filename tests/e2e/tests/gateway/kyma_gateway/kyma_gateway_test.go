package kyma_gateway

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "embed"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	apigatewayasserts "github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/gateway"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/asserts/resourceasserts"
	e2eclient "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	httpbinhelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpbin"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/v1access"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	checkTimeout          = 2 * time.Minute
	apiGatewayCRNamespace = "kyma-system"
	apiGatewayCRName      = "default"
)

//go:embed vs.yaml
var VirtualService string

var oathkeeperCRDResource = resourceasserts.StructCheck{Gvk: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"}, Name: "rules.oathkeeper.ory.sh", Namespace: ""}

var oathkeeperResources = []resourceasserts.StructCheck{
	{Gvk: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}, Name: "ory-oathkeeper", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, Name: "ory-oathkeeper-config", Namespace: apiGatewayCRNamespace},
	oathkeeperCRDResource,
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}, Name: "ory-oathkeeper-jwks-secret", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, Name: "ory-oathkeeper-api", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, Name: "ory-oathkeeper-proxy", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}, Name: "ory-oathkeeper-maester-metrics", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}, Name: "ory-oathkeeper", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}, Name: "oathkeeper-maester-account", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"}, Name: "oathkeeper-maester-role", Namespace: ""},
	{Gvk: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"}, Name: "oathkeeper-maester-role-binding", Namespace: ""},
	{Gvk: schema.GroupVersionKind{Group: "security.istio.io", Version: "v1beta1", Kind: "PeerAuthentication"}, Name: "ory-oathkeeper-maester-metrics", Namespace: apiGatewayCRNamespace},
	{Gvk: schema.GroupVersionKind{Group: "policy", Version: "v1", Kind: "PodDisruptionBudget"}, Name: "ory-oathkeeper", Namespace: apiGatewayCRNamespace},
}

func TestKymaGateway(t *testing.T) {
	t.Run("Oathkeeper is installed and uninstalled depending on APIGateway presence", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r, t))
		modulehelpers.SetupBaseCR(t)

		resourceasserts.AssertResourceExists(t, r, resourceasserts.StructCheck{Gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"}, Name: "kyma-gateway", Namespace: apiGatewayCRNamespace}, checkTimeout)
		require.NoError(t, wait.For(
			conditions.New(r).DeploymentAvailable("ory-oathkeeper", apiGatewayCRNamespace),
			wait.WithTimeout(2*time.Minute),
		), "Deployment %s/%s should be Ready", apiGatewayCRNamespace, "ory-oathkeeper")

		for _, rc := range oathkeeperResources {
			resourceasserts.AssertResourceExists(t, r, rc, checkTimeout)
		}

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))
		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)

		for _, rc := range oathkeeperResources {
			resourceasserts.AssertResourceDoesNotExist(t, r, rc, checkTimeout)
		}
	})

	t.Run("Oathkeeper is not installed when there is no api-gateway-config ConfigMap", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r, t))
		modulehelpers.SetupBaseCR(t)

		cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: v1access.V1AccessConfigMapName, Namespace: v1access.V1AccessConfigMapNamespace}}
		// CreateAllowAPIRuleV1Signatures registers a t.Cleanup to delete this ConfigMap. The manual delete
		// here tests behavior when the ConfigMap is deleted mid-run
		require.NoError(t, r.Delete(context.Background(), &cm))

		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}, Name: v1access.V1AccessConfigMapName, Namespace: v1access.V1AccessConfigMapNamespace}, checkTimeout)
		// The integration feature describes APIGateway removal first. In practice, without waiting for
		// asynchronous deletion to complete, the effective order is nondeterministic. This test uses a
		// deterministic sequence: remove the access ConfigMap first, then delete APIGateway with an
		// explicit wait, so the expected Oathkeeper Rule CRD state can be asserted reliably.
		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))
		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)

		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
		for _, rc := range oathkeeperResources {
			if rc.Gvk.Kind == oathkeeperCRDResource.Gvk.Kind && rc.Name == oathkeeperCRDResource.Name {
				continue
			}
			resourceasserts.AssertResourceDoesNotExist(t, r, rc, checkTimeout)
		}
		resourceasserts.AssertResourceExists(t, r, oathkeeperCRDResource, checkTimeout)
	})

	t.Run("Kyma Gateway is not removed when there is a VirtualService", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		svcName, _, err := httpbinhelpers.DeployHttpbin(t, apiGatewayCRNamespace)
		require.NoError(t, err, "Failed to deploy httpbin service")

		vsName := "kyma-vs-" + envconf.RandomName("", 6)
		vs, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			VirtualService,
			map[string]any{
				"Name":            vsName,
				"Namespace":       apiGatewayCRNamespace,
				"Host":            "local.kyma.dev",
				"Gateway":         apiGatewayCRNamespace + "/kyma-gateway",
				"DestinationHost": fmt.Sprintf("%s.%s.svc.cluster.local", svcName, apiGatewayCRNamespace),
			},
		)
		require.NoError(t, err, "Failed to create VirtualService resource")

		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		cr := operatorv1alpha1.APIGateway{ObjectMeta: metav1.ObjectMeta{Name: apiGatewayCRName, Namespace: apiGatewayCRNamespace}}
		require.NoError(t, r.Get(context.Background(), cr.GetName(), cr.GetNamespace(), &cr), "Failed to get APIGateway CR")
		cr.Spec.EnableKymaGateway = new(false)
		require.NoError(t, r.Update(context.Background(), &cr), "Failed to update APIGateway CR to disable Kyma Gateway")

		apigatewayasserts.AssertAPIGatewayCRWarning(t, r, apiGatewayCRNamespace, apiGatewayCRName)
		apigatewayasserts.AssertAPIGatewayCRDescriptionContains(t, r, apiGatewayCRNamespace, apiGatewayCRName, "There are custom resources that block the deletion of Kyma Gateway")

		require.NoError(t, r.Delete(context.Background(), vs), "Failed to delete VirtualService resource")

		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "VirtualService"}, Name: vs.GetName(), Namespace: vs.GetNamespace()}, checkTimeout)
		apigatewayasserts.AssertAPIGatewayCRReady(t, r, apiGatewayCRNamespace, apiGatewayCRName)
		apigatewayasserts.AssertAPIGatewayCRDescriptionContains(t, r, apiGatewayCRNamespace, apiGatewayCRName, "Successfully reconciled")
	})

	t.Run("Kyma Gateway is removed when there is no blocking resources", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))
		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)

		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.StructCheck{Gvk: schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"}, Name: "kyma-gateway", Namespace: apiGatewayCRNamespace}, checkTimeout)
	})

	t.Run("Second APIGateway CR is applied to the cluster", func(t *testing.T) {
		modulehelpers.SetupBaseCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		secondCR := &operatorv1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{Name: "second-api-gateway-cr", Namespace: apiGatewayCRNamespace},
		}
		require.NoError(t, r.Create(t.Context(), secondCR), "Failed to create second APIGateway CR")

		apigatewayasserts.AssertAPIGatewayCRWarning(t, r, secondCR.Namespace, secondCR.Name)
		apigatewayasserts.AssertAPIGatewayCRDescriptionContains(t, r, secondCR.Namespace, secondCR.Name, "stopped APIGateway CR reconciliation: only APIGateway CR default reconciles the module")
		apigatewayasserts.AssertAPIGatewayCRReady(t, r, apiGatewayCRNamespace, apiGatewayCRName)
		apigatewayasserts.AssertAPIGatewayCRDescriptionContains(t, r, apiGatewayCRNamespace, apiGatewayCRName, "Successfully reconciled")

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, secondCR.Name), "Failed to delete second APIGateway CR")
	})
}
