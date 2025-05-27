package clusterconfig

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterSize int

const (
	UnknownSize ClusterSize = iota
	Evaluation
	Production

	ProductionClusterCpuThreshold      int64 = 5
	ProductionClusterMemoryThresholdGi int64 = 10
)

func (s ClusterSize) String() string {
	switch s {
	case Evaluation:
		return "Evaluation"
	case Production:
		return "Production"
	default:
		return "Unknown"
	}
}

// EvaluateClusterSize counts the entire capacity of cpu and memory in the cluster and returns Evaluation
// if the total capacity of any of the resources is lower than ProductionClusterCpuThreshold or ProductionClusterMemoryThresholdGi.
func EvaluateClusterSize(ctx context.Context, k8sClient client.Client) (ClusterSize, error) {
	nodeList := corev1.NodeList{}
	err := k8sClient.List(ctx, &nodeList)
	if err != nil {
		return UnknownSize, err
	}

	var cpuCapacity resource.Quantity
	var memoryCapacity resource.Quantity
	for _, node := range nodeList.Items {
		nodeCpuCap := node.Status.Capacity.Cpu()
		if nodeCpuCap != nil {
			cpuCapacity.Add(*nodeCpuCap)
		}
		nodeMemoryCap := node.Status.Capacity.Memory()
		if nodeMemoryCap != nil {
			memoryCapacity.Add(*nodeMemoryCap)
		}
	}

	if cpuCapacity.Cmp(*resource.NewQuantity(ProductionClusterCpuThreshold, resource.DecimalSI)) == -1 ||
		memoryCapacity.Cmp(*resource.NewScaledQuantity(ProductionClusterMemoryThresholdGi, resource.Giga)) == -1 {
		return Evaluation, nil
	}
	return Production, nil
}
