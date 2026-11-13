package oathkeeper_test

import (
	"context"
	"github.com/kyma-project/api-gateway/internal/clusterconfig"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thoas/go-funk"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
)

var _ = Describe("Oathkeeper Deployment reconciliation", func() {
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

	It("Should deploy light version of Deployment on small cluster", func() {
		smallNode := corev1.Node{
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
		k8sClient := createFakeClient(&smallNode, apiGateway)
		status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
		Expect(status.IsReady()).To(BeTrue(), "%#v", status)

		var deployment appsv1.Deployment
		Expect(k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "ory-oathkeeper",
		}, &deployment)).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(2))
		oathkeeperContainter := funk.Find(deployment.Spec.Template.Spec.Containers, func(c corev1.Container) bool {
			return c.Name == "oathkeeper"
		}).(corev1.Container)

		oathkeeperMaesterContainter := funk.Find(deployment.Spec.Template.Spec.Containers, func(c corev1.Container) bool {
			return c.Name == "oathkeeper-maester"
		}).(corev1.Container)

		Expect(oathkeeperContainter.Resources.Limits.Cpu().String()).To(Equal("100m"))
		Expect(oathkeeperContainter.Resources.Limits.Memory().String()).To(Equal("128Mi"))
		Expect(oathkeeperContainter.Resources.Requests.Cpu().String()).To(Equal("10m"))
		Expect(oathkeeperContainter.Resources.Requests.Memory().String()).To(Equal("64Mi"))

		Expect(oathkeeperMaesterContainter.Resources.Limits.Cpu().String()).To(Equal("100m"))
		Expect(oathkeeperMaesterContainter.Resources.Limits.Memory().String()).To(Equal("50Mi"))
		Expect(oathkeeperMaesterContainter.Resources.Requests.Cpu().String()).To(Equal("10m"))
		Expect(oathkeeperMaesterContainter.Resources.Requests.Memory().String()).To(Equal("20Mi"))
	})

	It("Should deploy full version of Deployment on big cluster", func() {
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

		var deployment appsv1.Deployment
		Expect(k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "ory-oathkeeper",
		}, &deployment)).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(2))
		oathkeeperContainter := funk.Find(deployment.Spec.Template.Spec.Containers, func(c corev1.Container) bool {
			return c.Name == "oathkeeper"
		}).(corev1.Container)

		oathkeeperMaesterContainter := funk.Find(deployment.Spec.Template.Spec.Containers, func(c corev1.Container) bool {
			return c.Name == "oathkeeper-maester"
		}).(corev1.Container)

		Expect(oathkeeperContainter.Resources.Limits.Cpu().String()).To(Equal("10"))
		Expect(oathkeeperContainter.Resources.Limits.Memory().String()).To(Equal("512Mi"))
		Expect(oathkeeperContainter.Resources.Requests.Cpu().String()).To(Equal("100m"))
		Expect(oathkeeperContainter.Resources.Requests.Memory().String()).To(Equal("64Mi"))

		Expect(oathkeeperMaesterContainter.Resources.Limits.Cpu().String()).To(Equal("400m"))
		Expect(oathkeeperMaesterContainter.Resources.Limits.Memory().String()).To(Equal("1Gi"))
		Expect(oathkeeperMaesterContainter.Resources.Requests.Cpu().String()).To(Equal("10m"))
		Expect(oathkeeperMaesterContainter.Resources.Requests.Memory().String()).To(Equal("32Mi"))
	})

	It("Should not overwrite number of existing deployment replicas", func() {
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

		var initialReplicas = int32(4)
		initialDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ory-oathkeeper",
				Namespace: "kyma-system",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &initialReplicas,
			},
		}

		apiGateway := createApiGateway()
		k8sClient := createFakeClient(&node, apiGateway, &initialDeployment)
		status := oathkeeper.Reconcile(context.Background(), k8sClient, apiGateway)
		Expect(status.IsReady()).To(BeTrue(), "%#v", status)

		var deployment appsv1.Deployment
		Expect(k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "kyma-system",
			Name:      "ory-oathkeeper",
		}, &deployment)).Should(Succeed())

		Expect(*deployment.Spec.Replicas).To(Equal(initialReplicas))
	})
})
