package kyma_gateway

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "embed"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	e2eclient "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/client"
	httpbinhelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/httpbin"
	infrahelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/infrastructure"
	modulehelpers "github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/modules"
	"github.com/kyma-project/api-gateway/tests/e2e/pkg/helpers/v1access"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

const (
	checkTimeout  = 2 * time.Minute
	checkInterval = 2 * time.Second
)

//go:embed vs.yaml
var VirtualService string

type resourceCheck struct {
	gvk       schema.GroupVersionKind
	name      string
	namespace string
}

func setupBaselineCR(t *testing.T) {
	t.Helper()

	require.NoError(t, modulehelpers.CreateIstioOperatorCR(t))
	require.NoError(t, modulehelpers.CreateApiGatewayCR(t))
}

func deleteDefaultAPIGateway(t *testing.T, r *resources.Resources) {
	t.Helper()
	cr := &operatorv1alpha1.APIGateway{}
	cr.Name = "default"
	cr.Namespace = "kyma-system"

	err := r.Delete(t.Context(), cr)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		getErr := r.Get(context.Background(), cr.Name, cr.Namespace, cr)
		return k8serrors.IsNotFound(getErr)
	}, checkTimeout, checkInterval, "APIGateway CR should be deleted")
}
func assertResourceExists(t *testing.T, r *resources.Resources, rc resourceCheck) {
	t.Helper()
	require.Eventually(t, func() bool {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(rc.gvk)
		obj.SetName(rc.name)
		obj.SetNamespace(rc.namespace)
		return r.Get(context.Background(), rc.name, rc.namespace, obj) == nil
	}, checkTimeout, checkInterval, "%s %s/%s should exist", rc.gvk.Kind, rc.namespace, rc.name)
}

func assertResourceDoesNotExist(t *testing.T, r *resources.Resources, rc resourceCheck) {
	t.Helper()
	require.Eventually(t, func() bool {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(rc.gvk)
		obj.SetName(rc.name)
		obj.SetNamespace(rc.namespace)
		return k8serrors.IsNotFound(r.Get(context.Background(), rc.name, rc.namespace, obj))
	}, checkTimeout, checkInterval, "%s %s/%s should not exist", rc.gvk.Kind, rc.namespace, rc.name)
}

var oathkeeperResources = []resourceCheck{
	{
		gvk:       schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		name:      "ory-oathkeeper",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		name:      "ory-oathkeeper-config",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"},
		name:      "rules.oathkeeper.ory.sh",
		namespace: "",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		name:      "ory-oathkeeper-jwks-secret",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		name:      "ory-oathkeeper-api",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		name:      "ory-oathkeeper-proxy",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		name:      "ory-oathkeeper-maester-metrics",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		name:      "ory-oathkeeper",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		name:      "oathkeeper-maester-account",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		name:      "oathkeeper-maester-role",
		namespace: "",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
		name:      "oathkeeper-maester-role-binding",
		namespace: "",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "security.istio.io", Version: "v1beta1", Kind: "PeerAuthentication"},
		name:      "ory-oathkeeper-maester-metrics",
		namespace: "kyma-system",
	},
	{
		gvk:       schema.GroupVersionKind{Group: "policy", Version: "v1", Kind: "PodDisruptionBudget"},
		name:      "ory-oathkeeper",
		namespace: "kyma-system",
	},
}

func TestKymaGateway(t *testing.T) {
	t.Run("Oathkeeper is installed and uninstalled depending on APIGateway presence", func(t *testing.T) {
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r))
		setupBaselineCR(t)
		kymaGatewayResource := resourceCheck{
			gvk:       schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"},
			name:      "kyma-gateway",
			namespace: "kyma-system",
		}

		assertResourceExists(t, r, kymaGatewayResource)

		require.Eventually(t, func() bool {
			dep := &appsv1.Deployment{}
			if err := r.Get(context.Background(), "ory-oathkeeper", "kyma-system", dep); err != nil {
				return false
			}
			for _, condition := range dep.Status.Conditions {
				if condition.Type == appsv1.DeploymentAvailable && condition.Status == "True" {
					return true
				}
			}
			return false
		}, checkTimeout, checkInterval, "Deployment %s/%s should be Ready", "kyma-system", "ory-oathkeeper")

		for _, rc := range oathkeeperResources {
			assertResourceExists(t, r, rc)
		}

		deleteDefaultAPIGateway(t, r)

		for _, rc := range oathkeeperResources {
			assertResourceDoesNotExist(t, r, rc)
		}

	})
	t.Run("Oathkeeper is not installed when there is no api-gateway-config ConfigMap", func(t *testing.T) {

		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)
		require.NoError(t, v1access.CreateAllowAPIRuleV1Signatures(context.Background(), r))
		setupBaselineCR(t)
		deleteDefaultAPIGateway(t, r)

		cm := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      v1access.V1AccessConfigMapName,
				Namespace: v1access.V1AccessConfigMapNamespace,
			},
		}
		deleteErr := r.Delete(context.Background(), &cm)
		require.NoError(t, deleteErr)

		accessCm := resourceCheck{
			gvk:       schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
			name:      v1access.V1AccessConfigMapName,
			namespace: v1access.V1AccessConfigMapNamespace,
		}
		assertResourceDoesNotExist(t, r, accessCm)
		err = modulehelpers.CreateApiGatewayCR(t)
		require.NoError(t, err)

		for _, rc := range oathkeeperResources {
			assertResourceDoesNotExist(t, r, rc)
		} /*
			oathkeeperRuleCheck := resourceCheck{
				gvk:       schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1", Kind: "CustomResourceDefinition"},
				name:      "rules.oathkeeper.ory.sh",
				namespace: "",
			}
			assertResourceExists(t, r, oathkeeperRuleCheck)*/

	})
	t.Run("Kyma Gateway is not removed when there is a VirtualService", func(t *testing.T) {
		setupBaselineCR(t)
		svcName, _, err := httpbinhelpers.DeployHttpbin(t, "kyma-system")
		require.NoError(t, err, "Failed to deploy httpbin service")

		vsName := "kyma-vs"
		vs, err := infrahelpers.CreateResourceWithTemplateValues(
			t,
			VirtualService,
			map[string]any{
				"Name":            vsName,
				"Namespace":       "kyma-system",
				"Host":            "local.kyma.dev",
				"Gateway":         fmt.Sprintf("kyma-system/kyma-gateway"),
				"DestinationHost": fmt.Sprintf("%s.kyma-system.svc.cluster.local", svcName),
			},
		)

		require.NoError(t, err, "Failed to create VirtualService resource")

		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		cr := operatorv1alpha1.APIGateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: "kyma-system",
			},
		}
		getErr := r.Get(context.Background(), cr.GetName(), cr.GetNamespace(), &cr)
		require.NoError(t, getErr, "Failed to get APIGateway CR")
		cr.Spec.EnableKymaGateway = new(false)
		updateErr := r.Update(context.Background(), &cr)
		require.NoError(t, updateErr, "Failed to update APIGateway CR to disable Kyma Gateway")

		istioGW := &unstructured.Unstructured{}
		istioGW.SetGroupVersionKind(schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"})
		istioGW.SetName("kyma-gateway")
		istioGW.SetNamespace("kyma-system")
		err = r.Get(context.Background(), "kyma-gateway", "kyma-system", istioGW)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			getErr := r.Get(context.Background(), cr.GetName(), cr.GetNamespace(), &cr)
			return getErr == nil
		}, checkTimeout, checkInterval, "APIGateway CR should still exist while VirtualService is present")

		require.Eventually(t, func() bool {
			if getErr := r.Get(context.Background(), cr.GetName(), cr.GetNamespace(), &cr); getErr != nil {
				return false
			}
			return cr.Status.State == "Warning" &&
				cr.Status.Description == "There are custom resources that block the deletion of Kyma Gateway. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning"
		}, checkTimeout, checkInterval, "APIGateway CR should be in Warning state while VirtualService is present")

		err = r.Get(context.Background(), "kyma-gateway", "kyma-system", istioGW)
		require.NoError(t, err)

		deleteErr := r.Delete(context.Background(), vs)
		require.NoError(t, deleteErr, "Failed to delete VirtualService resource")

		require.Eventually(t, func() bool {
			getErr := r.Get(context.Background(), vs.GetName(), vs.GetNamespace(), vs)
			return k8serrors.IsNotFound(getErr)
		}, checkTimeout, checkInterval, "VirtualService should be removed")

		require.Eventually(t, func() bool {
			if getErr := r.Get(context.Background(), cr.GetName(), cr.GetNamespace(), &cr); getErr != nil {
				return false
			}
			return cr.Status.State == "Ready" && cr.Status.Description == "Successfully reconciled"
		}, checkTimeout, checkInterval, "APIGateway CR should be in Ready state after VirtualService is removed")
	})

	t.Run("Kyma Gateway is removed when there is no blocking resources", func(t *testing.T) {
		setupBaselineCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		deleteDefaultAPIGateway(t, r)

		gw := resourceCheck{
			gvk:       schema.GroupVersionKind{Group: "networking.istio.io", Version: "v1beta1", Kind: "Gateway"},
			name:      "kyma-gateway",
			namespace: "kyma-system",
		}
		assertResourceDoesNotExist(t, r, gw)
	})

	t.Run("Second APIGateway CR is applied to the cluster", func(t *testing.T) {
		setupBaselineCR(t)
		r, err := e2eclient.ResourcesClient(t)
		require.NoError(t, err)

		secondCR := &operatorv1alpha1.APIGateway{}
		secondCR.Name = "second-api-gateway-cr"
		secondCR.Namespace = "kyma-system"

		require.NoError(t, r.Create(t.Context(), secondCR), "Failed to create second APIGateway CR")

		require.Eventually(t, func() bool {
			created := &operatorv1alpha1.APIGateway{}
			if getErr := r.Get(context.Background(), secondCR.Name, secondCR.Namespace, created); getErr != nil {
				return false
			}
			return created.Status.State == "Warning" &&
				created.Status.Description == "stopped APIGateway CR reconciliation: only APIGateway CR default reconciles the module"
		}, checkTimeout, checkInterval, "Second APIGateway CR should be in Warning state")

		require.Eventually(t, func() bool {
			defaultCR := &operatorv1alpha1.APIGateway{}
			if getErr := r.Get(context.Background(), "default", "kyma-system", defaultCR); getErr != nil {
				return false
			}
			return defaultCR.Status.State == "Ready" && defaultCR.Status.Description == "Successfully reconciled"
		}, checkTimeout, checkInterval, "Default APIGateway CR should be in Ready state")

		require.NoError(t, r.Delete(t.Context(), secondCR), "Failed to delete second APIGateway CR")

		require.Eventually(t, func() bool {
			deleted := &operatorv1alpha1.APIGateway{}
			getErr := r.Get(context.Background(), secondCR.Name, secondCR.Namespace, deleted)
			return k8serrors.IsNotFound(getErr)
		}, checkTimeout, checkInterval, "Second APIGateway CR should be removed")
	})
}
