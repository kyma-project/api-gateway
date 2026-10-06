package oathkeeper_test

import (
	"context"
	"encoding/base64"
	"os"
	"time"

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

	crclient "sigs.k8s.io/controller-runtime/pkg/client"
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
			state, _, _, err := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(err).ToNot(BeNil())
			Expect(state).To(Equal(v1alpha1.Error))
		})

		It("Should successfully reconcile Oathkeeper", func() {
			apiGateway := createApiGateway()
			k8sClient := createFakeClient(apiGateway)
			status, _, _, err := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(v1alpha1.Ready), "%#v", status)

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
			status, _, _, err := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(v1alpha1.Ready), "%#v", status)
			apiGateway.DeletionTimestamp = &metav1.Time{Time: time.Now()}

			status, _, _, err = oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(v1alpha1.Ready), "%#v", status)

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
			status, desc, _, err := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(status).To(Equal(v1alpha1.Error))
			Expect(desc).To(Equal("Oathkeeper did not reconcile successfully"))
			Expect(err).To(HaveOccurred())
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
			status, _, _, err := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(v1alpha1.Ready))
		})
	})

	Context("ReconcileAndVerifyReadiness", func() {
		It("Should return error status with condition when reconciliation fails", func() {
			apiGateway := createApiGateway()
			deprecatedV1ConfigMap, err := apiruleAccessMaps()
			Expect(err).To(BeNil())

			k8sClient := createFakeClientThatFailsOnCreate(append([]crclient.Object{apiGateway}, deprecatedV1ConfigMap...)...)

			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}

			status, desc, cond, err := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)
			Expect(status).To(Equal(v1alpha1.Error))
			Expect(desc).To(Equal("Oathkeeper did not reconcile successfully"))
			Expect(err).To(HaveOccurred())
			Expect(cond).To(Not(BeNil()))
			Expect(cond.Type).To(Equal(v1alpha1.ConditionTypeReady))
			Expect(cond.Reason).To(Equal(v1alpha1.ReasonOathkeeperReconcileFailed))
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
		})

		It("Should return Ready status with condition for Oathkeeper deployment that is Available", func() {
			apiGateway := createApiGateway()
			deprecatedV1ConfigMap, err := apiruleAccessMaps()
			Expect(err).To(BeNil())

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

			k8sClient := createFakeClient(append([]crclient.Object{apiGateway, oathkeeperDep}, deprecatedV1ConfigMap...)...)
			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}
			status, _, cond, err := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(v1alpha1.Ready))
			Expect(cond).To(Not(BeNil()))
			Expect(cond.Type).To(Equal(v1alpha1.ConditionTypeReady))
			Expect(cond.Reason).To(Equal(v1alpha1.ReasonOathkeeperReconcileSucceeded))
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		})

		It("Should return Error for Oathkeeper deployment that is not Available", func() {
			apiGateway := createApiGateway()
			deprecatedV1ConfigMaps, err := apiruleAccessMaps()
			Expect(err).To(BeNil())

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

			k8sClient := createFakeClient(append([]crclient.Object{apiGateway, oathkeeperDep}, deprecatedV1ConfigMaps...)...)
			reconciler := oathkeeper.Reconciler{
				ReadinessRetryConfig: oathkeeper.RetryConfig{
					Attempts: 1,
					Delay:    1 * time.Millisecond,
				},
			}
			status, desc, _, err := reconciler.ReconcileAndVerifyReadiness(context.Background(), k8sClient, apiGateway)
			Expect(status).To(Equal(v1alpha1.Error))
			Expect(desc).To(Equal("Oathkeeper did not start successfully"))
			Expect(err).To(HaveOccurred())
		})

	})

})

func createApiGateway() *v1alpha1.APIGateway {
	return &v1alpha1.APIGateway{
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.APIGatewaySpec{},
	}
}

func apiruleAccessMaps() ([]crclient.Object, error) {
	data, err := base64.StdEncoding.DecodeString("xEYGAAobIJRdbtfrgZYkBehKLGT3pI8YVu22FPHyHJWVjpTzvSPa+8vQFjsiHcrLvmDfEy56Y/D9Xfq/Qtt6o41bvKMqJPUByxRiAAAAAABsb2NhbC5reW1hLmRldsKYBgAbCgAAACkFgmj7jOoioQb7y9AWOyIdysu+YN8TLnpj8P1d+r9C23qjjVu8oyok9QAAAACp7CCUXW7X64GWJAXoSixk96SPGFbtthTx8hyVlY6U870j2t8v/C1gL5Vkw9+y7sfd/GKzAZGIwlf6+XDM8U4VlHtS/CRKP155fLX9g96/jixWU7JZgCf3Yo/a5Bwjg0TYkQM=")
	if err != nil {
		return nil, err
	}
	return []crclient.Object{
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "shoot-info",
				Namespace: "kube-system",
			},
			Data: map[string]string{
				"domain": "local.kyma.dev",
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "apirule-access",
				Namespace: "kyma-system",
			},
			BinaryData: map[string][]byte{
				"access.sig": data,
			},
		},
	}, nil
}
