package clusterconfig_test

import (
	"context"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/clusterconfig"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("EvaluateClusterSize", func() {
	It("should return Evaluation when cpu capacity is less than ProductionClusterCpuThreshold", func() {
		//given
		k3dNode := corev1.Node{
			ObjectMeta: v1.ObjectMeta{
				Name: "k3d-node-1",
			},
			Status: corev1.NodeStatus{
				Capacity: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewQuantity(clusterconfig.ProductionClusterCpuThreshold-1, resource.DecimalSI),
					"memory": *resource.NewScaledQuantity(int64(32), resource.Giga),
				},
			},
		}

		client := createFakeClient(&k3dNode)

		//when
		size, err := clusterconfig.EvaluateClusterSize(context.Background(), client)

		//then
		Expect(err).To(Not(HaveOccurred()))
		Expect(size).To(Equal(clusterconfig.Evaluation))
	})

	It("should return Evaluation when memory capacity is less than ProductionClusterMemoryThresholdGi", func() {
		//given
		k3dNode := corev1.Node{
			ObjectMeta: v1.ObjectMeta{
				Name: "k3d-node-1",
			},
			Status: corev1.NodeStatus{
				Capacity: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewMilliQuantity(int64(12000), resource.DecimalSI),
					"memory": *resource.NewScaledQuantity(clusterconfig.ProductionClusterMemoryThresholdGi-1, resource.Giga),
				},
			},
		}

		client := createFakeClient(&k3dNode)

		//when
		size, err := clusterconfig.EvaluateClusterSize(context.Background(), client)

		//then
		Expect(err).To(Not(HaveOccurred()))
		Expect(size).To(Equal(clusterconfig.Evaluation))
	})

	It("should return Production when memory capacity is bigger than ProductionClusterMemoryThresholdGi and CPU capacity is bigger than ProductionClusterCpuThreshold", func() {
		//given
		k3dNode := corev1.Node{
			ObjectMeta: v1.ObjectMeta{
				Name: "k3d-node-1",
			},
			Status: corev1.NodeStatus{
				Capacity: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewQuantity(clusterconfig.ProductionClusterCpuThreshold, resource.DecimalSI),
					"memory": *resource.NewScaledQuantity(clusterconfig.ProductionClusterMemoryThresholdGi, resource.Giga),
				},
			},
		}

		k3dNode2 := corev1.Node{
			ObjectMeta: v1.ObjectMeta{
				Name: "k3d-node-2",
			},
			Status: corev1.NodeStatus{
				Capacity: map[corev1.ResourceName]resource.Quantity{
					"cpu":    *resource.NewQuantity(clusterconfig.ProductionClusterCpuThreshold, resource.DecimalSI),
					"memory": *resource.NewScaledQuantity(clusterconfig.ProductionClusterMemoryThresholdGi, resource.Giga),
				},
			},
		}

		client := createFakeClient(&k3dNode, &k3dNode2)

		//when
		size, err := clusterconfig.EvaluateClusterSize(context.Background(), client)

		//then
		Expect(err).To(Not(HaveOccurred()))
		Expect(size).To(Equal(clusterconfig.Production))
	})
})

func createFakeClient(objects ...client.Object) client.Client {
	err := operatorv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).ShouldNot(HaveOccurred())
	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).ShouldNot(HaveOccurred())

	return fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).Build()
}
