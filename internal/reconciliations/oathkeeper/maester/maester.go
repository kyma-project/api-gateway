package maester

import (
	"context"
	"errors"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileMaester(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {
	err := errors.Join(
		reconcileOryOathkeeperPeerAuthentication(ctx, k8sClient, apiGatewayCR),
		reconcileOryOathkeeperMaesterServiceAccount(ctx, k8sClient, apiGatewayCR),
	)
	var clusterRoleErr error
	if !apiGatewayCR.IsInDeletion() {
		clusterRoleErr = errors.Join(
			reconcileOryOathkeeperMaesterClusterRole(ctx, k8sClient, apiGatewayCR),
			reconcileOryOathkeeperMaesterClusterRoleBinding(ctx, k8sClient, apiGatewayCR),
		)
	} else {
		clusterRoleErr = errors.Join(
			reconcileOryOathkeeperMaesterClusterRoleBinding(ctx, k8sClient, apiGatewayCR),
			reconcileOryOathkeeperMaesterClusterRole(ctx, k8sClient, apiGatewayCR),
		)
	}
	return errors.Join(err, clusterRoleErr)
}
