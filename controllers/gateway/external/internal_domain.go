package external

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nonGardenerDomainName = "local.kyma.dev"
)

// getGardenerShootInfo reads the shoot-info ConfigMap from kube-system
func getGardenerShootInfo(ctx context.Context, k8sClient client.Client) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	if err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      "shoot-info",
		Namespace: "kube-system",
	}, cm); err != nil {
		return nil, err
	}
	return cm, nil
}

// getGardenerDomain retrieves the cluster domain from Gardener shoot-info ConfigMap
func getGardenerDomain(ctx context.Context, k8sClient client.Client) (string, error) {
	cm, err := getGardenerShootInfo(ctx, k8sClient)
	if err != nil {
		return "", err
	}
	domain, exists := cm.Data["domain"]
	if !exists || domain == "" {
		return "", fmt.Errorf("domain not found in shoot-info ConfigMap")
	}
	return domain, nil
}

// buildInternalDomain constructs the full internal domain from subdomain and cluster domain
// If Gardener is available, it uses the Gardener domain. Otherwise, it falls back to nonGardenerDomainName
func (r *ExternalGatewayReconciler) buildInternalDomain(ctx context.Context, subdomain string) (string, error) {
	clusterDomain, err := getGardenerDomain(ctx, r.Client)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Gardener shoot-info not found, use fallback domain
			clusterDomain = nonGardenerDomainName
		} else {
			// Other error occurred
			return "", fmt.Errorf("failed to get Gardener cluster domain: %w", err)
		}
	}

	// If domain is empty, use fallback
	if clusterDomain == "" {
		clusterDomain = nonGardenerDomainName
	}

	return fmt.Sprintf("%s.%s", subdomain, clusterDomain), nil
}
