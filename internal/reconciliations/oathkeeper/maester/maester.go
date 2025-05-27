package maester

import (
	"context"
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
)

func ReconcileMaester(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	return errors.Join(
		reconcileOryOathkeeperPeerAuthentication(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperMaesterServiceAccount(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperMaesterClusterRole(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperMaesterClusterRoleBinding(ctx, k8sClient, apiGatewayCR),
	)
}
