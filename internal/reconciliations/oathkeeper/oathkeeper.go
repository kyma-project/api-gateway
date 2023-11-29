package oathkeeper

import (
	"context"
	"errors"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/cronjob"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/maester"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewReconciler() Reconciler {
	return Reconciler{}
}

type Reconciler struct {
}

func (r Reconciler) ReconcileAndVerifyReadiness(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) controllers.Status {
	status := Reconcile(ctx, k8sClient, apiGatewayCR)
	ctrl.Log.Info("Reconciled Oathkeeper", "status", status)
	if !status.IsReady() {
		return status
	}

	ctrl.Log.Info("Waiting for Oathkeeper Deployment to become ready")
	err := waitForOathkeeperDeploymentToBeReady(ctx, k8sClient)
	if err != nil {
		return controllers.ErrorStatus(err, "Oathkeeper did not start successfully")
	}

	return controllers.ReadyStatus()
}

func Reconcile(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) controllers.Status {
	err := reconcileOryOathkeeperRuleCRD(ctx, k8sClient, *apiGatewayCR)
	if err != nil {
		return controllers.ErrorStatus(err, "Oathkeeper Rule CRD did not reconcile successfully")
	}
	err = maester.ReconcileMaester(ctx, k8sClient, *apiGatewayCR)
	if err != nil {
		return controllers.ErrorStatus(err, "Oathkeeper Maester did not reconcile successfully")
	}

	err = errors.Join(
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
