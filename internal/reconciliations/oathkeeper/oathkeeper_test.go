package oathkeeper_test

import (
	"context"
	"os"
	"time"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/conditions"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type deployedResource struct {
	name       string
	namespaced bool
	GVK        schema.GroupVersionKind
}

var resourceList = []deployedResource{
	{
		name:       "ory-oathkeeper-api",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Service",
		},
	},
	{
		name:       "ory-oathkeeper-maester-metrics",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Service",
		},
	},
	{
		name:       "ory-oathkeeper-proxy",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Service",
		},
	},
	{
		name:       "ory-oathkeeper",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
	},
	{
		name:       "rules.oathkeeper.ory.sh",
		namespaced: false,
		GVK: schema.GroupVersionKind{
			Group:   "apiextensions.k8s.io",
			Version: "v1",
			Kind:    "CustomResourceDefinition",
		},
	},
	{
		name:       "ory-oathkeeper-config",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ConfigMap",
		},
	},
	{
		name:       "ory-oathkeeper",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ServiceAccount",
		},
	},
	{
		name:       "ory-oathkeeper-jwks-secret",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Secret",
		},
	},
	{
		name:       "ory-oathkeeper-maester-metrics",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "security.istio.io",
			Version: "v1beta1",
			Kind:    "PeerAuthentication",
		},
	},
	{
		name:       "oathkeeper-maester-account",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "ServiceAccount",
		},
	},
	{
		name:       "oathkeeper-maester-role-binding",
		namespaced: false,
		GVK: schema.GroupVersionKind{
			Group:   "rbac.authorization.k8s.io",
			Version: "v1",
			Kind:    "ClusterRoleBinding",
		},
	},
	{
		name:       "oathkeeper-maester-role",
		namespaced: false,
		GVK: schema.GroupVersionKind{
			Group:   "rbac.authorization.k8s.io",
			Version: "v1",
			Kind:    "ClusterRole",
		},
	},
	{
		name:       "ory-oathkeeper-maester-metrics",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "security.istio.io",
			Version: "v1beta1",
			Kind:    "PeerAuthentication",
		},
	},
	{
		name:       "ory-oathkeeper",
		namespaced: true,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "policy/v1",
			Kind:    "PodDisruptionBudget",
		},
	},
}

var _ = Describe("Oathkeeper reconciliation with environment not set", func() {
})

var _ = Describe("Oathkeeper reconciliation", func() {
	BeforeEach(func() {
		Expect(os.Setenv("oathkeeper", "oathkeeper:latest")).To(Succeed())
		Expect(os.Setenv("oathkeeper-maester", "oathkeeper:latest")).To(Succeed())
		Expect(os.Setenv("busybox", "busybox:latest")).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Unsetenv("oathkeeper")).To(Succeed())
		Expect(os.Unsetenv("oathkeeper-maester")).To(Succeed())
		Expect(os.Unsetenv("busybox")).To(Succeed())
	})

	Context("Reconcile", func() {
		It("Should fail if images are not set in environment variables", func() {
			Expect(os.Unsetenv("oathkeeper")).To(Succeed())
			Expect(os.Unsetenv("oathkeeper-maester")).To(Succeed())
			Expect(os.Unsetenv("busybox")).To(Succeed())

			apiGateway := createApiGateway()
			k8sClient := createFakeClient(apiGateway)
			status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status.IsError()).To(BeTrue(), "%#v", status)
		})

		It("Should successfully reconcile Oathkeeper", func() {
			apiGateway := createApiGateway()
			k8sClient := createFakeClient(apiGateway)
			status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status.IsReady()).To(BeTrue(), "%#v", status)

			for _, resource := range resourceList {
				var obj unstructured.Unstructured
				obj.SetGroupVersionKind(resource.GVK)
				var err error
				if resource.namespaced {
					err = k8sClient.Get(context.Background(), types.NamespacedName{
						Namespace: reconciliations.Namespace,
						Name:      resource.name,
					}, &obj)
				} else {
					err = k8sClient.Get(context.Background(), types.NamespacedName{
						Name: resource.name,
					}, &obj)
				}

				Expect(err).ShouldNot(HaveOccurred())
				Expect(obj.GetAnnotations()).To(HaveKeyWithValue("apigateways.operator.kyma-project.io/managed-by-disclaimer",
					"DO NOT EDIT - This resource is managed by Kyma.\nAny modifications are discarded and the resource is reverted to the original state."))

				Expect(obj.GetLabels()).To(HaveKeyWithValue("kyma-project.io/module", "api-gateway"))
			}
		})

		It("Should remove Oathkeeper resources on deletion", func() {
			apiGateway := createApiGateway()
			k8sClient := createFakeClient(apiGateway)
			status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status.IsReady()).To(BeTrue(), "%#v", status.NestedError())

			apiGateway.DeletionTimestamp = &metav1.Time{Time: time.Now()}

			status = oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status.IsReady()).To(BeTrue(), "%#v", status.NestedError())

			for _, resource := range resourceList {
				var obj unstructured.Unstructured
				obj.SetGroupVersionKind(resource.GVK)
				var err error
				if resource.namespaced {
					err = k8sClient.Get(context.Background(), types.NamespacedName{
						Namespace: reconciliations.Namespace,
						Name:      resource.name,
					}, &obj)
				} else {
					err = k8sClient.Get(context.Background(), types.NamespacedName{
						Name: resource.name,
					}, &obj)
				}

				Expect(k8serrors.IsNotFound(err)).To(BeTrue())
			}
		})

		It("Should return error status when reconciliation fails", func() {
			apiGateway := createApiGateway()
			k8sClient := createFakeClientThatFailsOnCreate()
			status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status.IsError()).To(BeTrue(), "%#v", status)
			Expect(status.Description()).To(Equal("Oathkeeper did not reconcile successfully"))
		})

		It("Should not fail when Gardener shoot-info without domain exists", func() {
			apiGateway := createApiGateway()
			cm := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "shoot-info",
					Namespace: "kube-system",
				},
				Data: map[string]string{},
			}
			k8sClient := createFakeClient(apiGateway, &cm)
			status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status.IsReady()).To(BeTrue(), "%#v", status)
		})
	})

	Context("ReconcileAndVerifyReadiness", func() {
		It("Should return error status with condition when reconciliation fails", func() {
			apiGateway := createApiGateway()
			deprecatedV1ConfigMap := createDeprecatedV1ConfigMap()
			k8sClient := createFakeClientThatFailsOnCreate(deprecatedV1ConfigMap)

			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}

			status := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)

			Expect(status.IsError()).To(BeTrue(), "%#v", status)
			Expect(status.Description()).To(Equal("Oathkeeper did not reconcile successfully"))
			Expect(status.Condition()).To(Not(BeNil()))
			Expect(status.Condition().Type).To(Equal(conditions.OathkeeperReconcileFailed.Condition().Type))
			Expect(status.Condition().Reason).To(Equal(conditions.OathkeeperReconcileFailed.Condition().Reason))
			Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
		})

		It("Should return Ready status with condition for Oathkeeper deployment that is Available", func() {
			apiGateway := createApiGateway()
			deprecatedV1ConfigMap := createDeprecatedV1ConfigMap()

			oathkeeperDep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ory-oathkeeper",
					Namespace: reconciliations.Namespace,
				},
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:   appsv1.DeploymentAvailable,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			k8sClient := createFakeClient(apiGateway, oathkeeperDep, deprecatedV1ConfigMap)
			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}
			status := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)
			Expect(status.IsReady()).To(BeTrue(), "%#v", status)
			Expect(status.Condition()).To(Not(BeNil()))
			Expect(status.Condition().Type).To(Equal(conditions.OathkeeperReconcileSucceeded.Condition().Type))
			Expect(status.Condition().Reason).To(Equal(conditions.OathkeeperReconcileSucceeded.Condition().Reason))
			Expect(status.Condition().Status).To(Equal(metav1.ConditionFalse))
		})

		It("Should return Error for Oathkeeper deployment that is not Available", func() {
			apiGateway := createApiGateway()
			deprecatedV1ConfigMap := createDeprecatedV1ConfigMap()

			oathkeeperDep := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ory-oathkeeper",
					Namespace: reconciliations.Namespace,
				},
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{
							Type:   appsv1.DeploymentAvailable,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}

			k8sClient := createFakeClient(apiGateway, oathkeeperDep, deprecatedV1ConfigMap)
			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}
			status := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)
			Expect(status.IsError()).To(BeTrue(), "%#v", status)
			Expect(status.Description()).To(Equal("Oathkeeper did not start successfully"))
		})

	})

})

func createApiGateway() *v1alpha1.APIGateway {
	return &v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.APIGatewaySpec{},
	}
}

func createDeprecatedV1ConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-gateway-config.operator.kyma-project.io",
			Namespace: "kyma-system",
		},
		Data: map[string]string{
			"enableDeprecatedV1beta1APIRule": "true",
		},
	}
}
