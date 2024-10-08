package reconciliations

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetGardenerDomain returns the domain name from the Gardener shoot-info config.
func GetGardenerDomain(ctx context.Context, k8sClient client.Client) (string, error) {
	cm, err := getGardenerShootInfo(ctx, k8sClient)
	if err != nil {
		return "", err
	}

	return cm.Data["domain"], nil
}

func getGardenerShootInfo(ctx context.Context, k8sClient client.Client) (corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{}
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: "kube-system", Name: "shoot-info"}, &cm)
	return cm, err
}
