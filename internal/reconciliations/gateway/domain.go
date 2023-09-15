package gateway

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const nonGardenerDomainName = "local.kyma.dev"

// getKymaGatewayDomain returns the domain name from the Gardener shoot-info config or a default domain name for other clusters.
func getKymaGatewayDomain(ctx context.Context, k8sClient client.Client) (string, error) {

	cm := corev1.ConfigMap{}
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: "kube-system", Name: "shoot-info"}, &cm)

	if err != nil && !k8serrors.IsNotFound(err) {
		return "", err
	}

	if err != nil && k8serrors.IsNotFound(err) {
		return nonGardenerDomainName, nil
	}

	if _, ok := cm.Data["domain"]; !ok {
		return "", fmt.Errorf("domain not found in Gardener shoot-info")
	}

	return cm.Data["domain"], nil
}
