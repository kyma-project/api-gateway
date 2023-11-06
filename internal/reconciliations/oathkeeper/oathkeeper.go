package oathkeeper

import (
	"context"
	"errors"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/cronjob"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/maester"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileOathkeeper(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) controllers.Status {
	err := errors.Join(
		reconcileOryOathkeeperRuleCRD(ctx, k8sClient, *apiGatewayCR),
		maester.ReconcileMaester(ctx, k8sClient, *apiGatewayCR),
		reconcileOryJWKSSecret(ctx, k8sClient, *apiGatewayCR),
		reconcileOryOathkeeperConfigConfigMap(ctx, k8sClient, *apiGatewayCR),
		reconcileOathkeeperHPA(ctx, k8sClient, *apiGatewayCR),
		reconcileOryOathkeeperServiceAccount(ctx, k8sClient, *apiGatewayCR),
		reconcileOryOathkeeperServices(ctx, k8sClient, *apiGatewayCR),
		cronjob.ReconcileCronjob(ctx, k8sClient, *apiGatewayCR),
		reconcileOathkeeperDeployment(ctx, k8sClient, *apiGatewayCR),
	)
	if err != nil {
		return controllers.ErrorStatus(err, "Oathkeeper did not reconcile successfully")
	}

	return controllers.ReadyStatus()
}
