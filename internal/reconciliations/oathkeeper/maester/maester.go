package maester

import (
	"context"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileMaester(ctx context.Context, k8sClient client.Client, apiGatewayCR v1alpha1.APIGateway) error {

	err := reconcileOryOathkeeperPeerAuthentication(ctx, k8sClient, apiGatewayCR)
	if err != nil {
		return err
	}
	err = reconcileOryOathkeeperMaesterServiceAccount(ctx, k8sClient, apiGatewayCR)
	if err != nil {
		return err
	}
	err = reconcileOryOathkeeperMaesterClusterRoleBinding(ctx, k8sClient, apiGatewayCR)
	if err != nil {
		return err
	}
	err = reconcileOryOathkeeperMaesterClusterRole(ctx, k8sClient, apiGatewayCR)
	if err != nil {
		return err
	}

	return nil
}
