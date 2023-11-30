package oathkeeper_test

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
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
	"time"
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
}

var _ = Describe("Oathkeeper reconciliation", func() {

	Context("Reconcile", func() {
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

		It("Should remove Cronjob resources from the cluster if those exist", func() {
			cronjobResources := []deployedResource{
				{
					name:       "oathkeeper-jwks-rotator",
					namespaced: true,
					GVK: schema.GroupVersionKind{
						Group:   "batch",
						Version: "v1",
						Kind:    "Cronjob",
					},
				},
				{
					name:       "ory-oathkeeper-keys-job-role",
					namespaced: true,
					GVK: schema.GroupVersionKind{
						Group:   "rbac.authorization.k8s.io",
						Version: "v1",
						Kind:    "Role",
					},
				},
				{
					name:       "ory-oathkeeper-keys-job-role-binding",
					namespaced: true,
					GVK: schema.GroupVersionKind{
						Group:   "rbac.authorization.k8s.io",
						Version: "v1",
						Kind:    "RoleBinding",
					},
				},
				{
					name:       "ory-oathkeeper-keys-service-account",
					namespaced: true,
					GVK: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "ServiceAccount",
					},
				},
			}

			apiGateway := createApiGateway()
			k8sClient := createFakeClient(apiGateway)

			for _, res := range cronjobResources {
				var r unstructured.Unstructured
				r.SetGroupVersionKind(res.GVK)
				r.SetName(res.name)
				if res.namespaced {
					r.SetNamespace(reconciliations.Namespace)
				}
				Expect(k8sClient.Create(context.Background(), &r)).To(Succeed())
			}

			status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status.IsReady()).To(BeTrue(), "%#v", status.NestedError())

			for _, resource := range cronjobResources {
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
	})

	Context("Reconciler", func() {

		It("Should return Ready for Oathkeeper deployment that is Available", func() {
			apiGateway := createApiGateway()

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

			k8sClient := createFakeClient(apiGateway, oathkeeperDep)
			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}
			status := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)
			Expect(status.IsReady()).To(BeTrue(), "%#v", status)
		})

		It("Should return Error for Oathkeeper deployment that is not Available", func() {
			apiGateway := createApiGateway()

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

			k8sClient := createFakeClient(apiGateway, oathkeeperDep)
			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}
			status := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)
			Expect(status.IsError()).To(BeTrue(), "%#v", status)
		})

	})

})

func createApiGateway() *v1alpha1.APIGateway {
	return &v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.APIGatewaySpec{},
	}
}
