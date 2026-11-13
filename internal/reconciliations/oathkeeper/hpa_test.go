package oathkeeper_test

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/clusterconfig"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"time"
)

var _ = Describe("Oathkeeper HPA reconciliation", func() {
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

	It("Should not apply HPA on small cluster", func() {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
			Status: corev1.NodeStatus{
				Capacity: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewQuantity(clusterconfig.ProductionClusterCpuThreshold-1, resource.DecimalSI),
					"memory": *resource.NewScaledQuantity(int64(32), resource.Giga),
				},
			},
		}

		apiGateway := createApiGateway()
		k8sClient := createFakeClient(&node, apiGateway)
		status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
		Expect(status.IsReady()).To(BeTrue(), "%#v", status)

		var hpa autoscalingv2.HorizontalPodAutoscaler
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "ory-oathkeeper",
		}, &hpa)

		Expect(err).To(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

	It("Should apply HPA on big cluster", func() {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
			Status: corev1.NodeStatus{
				Capacity: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewQuantity(clusterconfig.ProductionClusterCpuThreshold+1, resource.DecimalSI),
					"memory": *resource.NewScaledQuantity(int64(32), resource.Giga),
				},
			},
		}

		apiGateway := createApiGateway()
		k8sClient := createFakeClient(&node, apiGateway)
		status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
		Expect(status.IsReady()).To(BeTrue(), "%#v", status)

		var hpa autoscalingv2.HorizontalPodAutoscaler
		Expect(k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "ory-oathkeeper",
		}, &hpa)).Should(Succeed())
	})

	It("Should delete HPA when APIGateway is deleted", func() {
		node := corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
			Status: corev1.NodeStatus{
				Capacity: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewQuantity(clusterconfig.ProductionClusterCpuThreshold+1, resource.DecimalSI),
					"memory": *resource.NewScaledQuantity(int64(32), resource.Giga),
				},
			},
		}

		var initialHpa = autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ory-oathkeeper",
				Namespace: "kyma-system",
			},
		}

		apiGateway := createApiGateway()
		apiGateway.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		apiGateway.Finalizers = []string{"test"}

		k8sClient := createFakeClient(&node, apiGateway, &initialHpa)
		status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
		Expect(status.IsReady()).To(BeTrue(), "%#v", status)

		var hpa autoscalingv2.HorizontalPodAutoscaler
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "ory-oathkeeper",
		}, &hpa)

		Expect(err).To(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})

})
