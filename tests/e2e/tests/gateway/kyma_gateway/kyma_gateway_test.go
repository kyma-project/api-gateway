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

var oathkeeperResources = []resourceasserts.StructCheck{
	resourceasserts.Resource("apps", "v1", "Deployment", "ory-oathkeeper", apiGatewayCRNamespace),
	resourceasserts.Resource("", "v1", "ConfigMap", "ory-oathkeeper-config", apiGatewayCRNamespace),
	resourceasserts.Resource("apiextensions.k8s.io", "v1", "CustomResourceDefinition", "rules.oathkeeper.ory.sh", ""),
	resourceasserts.Resource("", "v1", "Secret", "ory-oathkeeper-jwks-secret", apiGatewayCRNamespace),
	resourceasserts.Resource("", "v1", "Service", "ory-oathkeeper-api", apiGatewayCRNamespace),
	resourceasserts.Resource("", "v1", "Service", "ory-oathkeeper-proxy", apiGatewayCRNamespace),
	resourceasserts.Resource("", "v1", "Service", "ory-oathkeeper-maester-metrics", apiGatewayCRNamespace),
	resourceasserts.Resource("", "v1", "ServiceAccount", "ory-oathkeeper", apiGatewayCRNamespace),
	resourceasserts.Resource("", "v1", "ServiceAccount", "oathkeeper-maester-account", apiGatewayCRNamespace),
	resourceasserts.Resource("rbac.authorization.k8s.io", "v1", "ClusterRole", "oathkeeper-maester-role", ""),
	resourceasserts.Resource("rbac.authorization.k8s.io", "v1", "ClusterRoleBinding", "oathkeeper-maester-role-binding", ""),
	resourceasserts.Resource("security.istio.io", "v1beta1", "PeerAuthentication", "ory-oathkeeper-maester-metrics", apiGatewayCRNamespace),
	resourceasserts.Resource("policy", "v1", "PodDisruptionBudget", "ory-oathkeeper", apiGatewayCRNamespace),
}

func TestKymaGateway(t *testing.T) {
	t.Run("Oathkeeper is installed and uninstalled depending on APIGateway presence", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r, t))
		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

		resourceasserts.AssertResourceExists(t, r, resourceasserts.Resource("networking.istio.io", "v1beta1", "Gateway", "kyma-gateway", apiGatewayCRNamespace), checkTimeout)
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
		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))

		cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: v1access.V1AccessConfigMapName, Namespace: v1access.V1AccessConfigMapNamespace}}
		// CreateAllowAPIRuleV1Signatures registers a t.Cleanup to delete this ConfigMap. The manual delete
		// here tests behavior when the ConfigMap is deleted mid-run
		require.NoError(t, r.Delete(context.Background(), &cm))

		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.Resource("", "v1", "ConfigMap", v1access.V1AccessConfigMapName, v1access.V1AccessConfigMapNamespace), checkTimeout)
		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))
		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)

		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
		for _, rc := range oathkeeperResources {
			resourceasserts.AssertResourceDoesNotExist(t, r, rc, checkTimeout)
		}
	})

	t.Run("Kyma Gateway is not removed when there is a VirtualService", func(t *testing.T) {
		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
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

		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.Resource("networking.istio.io", "v1beta1", "VirtualService", vs.GetName(), vs.GetNamespace()), checkTimeout)
		apigatewayasserts.AssertAPIGatewayCRReady(t, r, apiGatewayCRNamespace, apiGatewayCRName)
		apigatewayasserts.AssertAPIGatewayCRDescriptionContains(t, r, apiGatewayCRNamespace, apiGatewayCRName, "Successfully reconciled")
	})

	t.Run("Kyma Gateway is removed when there is no blocking resources", func(t *testing.T) {
		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		require.NoError(t, modulehelpers.DeleteAPIGateway(t, r, apiGatewayCRNamespace, apiGatewayCRName))
		apigatewayasserts.AssertAPIGatewayCRDeleted(t, r, apiGatewayCRNamespace, apiGatewayCRName)

		resourceasserts.AssertResourceDoesNotExist(t, r, resourceasserts.Resource("networking.istio.io", "v1beta1", "Gateway", "kyma-gateway", apiGatewayCRNamespace), checkTimeout)
	})

	t.Run("Second APIGateway CR is applied to the cluster", func(t *testing.T) {
		require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
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
