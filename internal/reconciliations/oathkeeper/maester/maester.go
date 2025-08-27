package maester

import (
	"context"
	"errors"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileMaester(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	return errors.Join(
		reconcileOryOathkeeperPeerAuthentication(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperMaesterServiceAccount(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperMaesterClusterRole(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperMaesterClusterRoleBinding(ctx, k8sClient, apiGatewayCR),
	)
}

func DeleteMaester(ctx context.Context, k8sClient client.Client) error {
	return errors.Join(
		deletePeerAuthentication(ctx, k8sClient, peerAuthenticationName, reconciliations.Namespace),
		deleteServiceAccount(ctx, k8sClient, ServiceAccountName, reconciliations.Namespace),
		deleteClusterRole(ctx, k8sClient, roleName),
		deleteRoleBinding(ctx, k8sClient, roleBindingName),
	)
}
