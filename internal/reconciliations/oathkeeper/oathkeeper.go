package oathkeeper

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/conditions"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/maester"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewReconciler() Reconciler {
	return Reconciler{
		ReadinessRetryConfig: RetryConfig{
			Attempts: 60,
			Delay:    2 * time.Second,
		},
	}
}

type Reconciler struct {
	ReadinessRetryConfig RetryConfig
}

type RetryConfig struct {
	Attempts uint
	Delay    time.Duration
}

func (r Reconciler) ReconcileAndVerifyReadiness(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) controllers.Status {
	status := Reconcile(ctx, k8sClient, apiGatewayCR)
	ctrl.Log.Info("Reconciled Oathkeeper", "status", status)
	if !status.IsReady() {
		return status
	}

	if !apiGatewayCR.IsInDeletion() {
		ctrl.Log.Info("Waiting for Oathkeeper Deployment to become ready")
		err := waitForOathkeeperDeploymentToBeReady(ctx, k8sClient, r.ReadinessRetryConfig)
		if err != nil {
			return controllers.ErrorStatus(err, "Oathkeeper did not start successfully", conditions.OathkeeperReconcileFailed.Condition())
		}
	}

	return controllers.ReadyStatus(conditions.OathkeeperReconcileSucceeded.Condition())
}

func Reconcile(ctx context.Context, k8sClient client.Client, apiGatewayCR *v1alpha1.APIGateway) controllers.Status {
	err := errors.Join(
		reconcileOryOathkeeperRuleCRD(ctx, k8sClient, *apiGatewayCR),
		maester.ReconcileMaester(ctx, k8sClient, *apiGatewayCR),
		reconcileOryJWKSSecret(ctx, k8sClient, *apiGatewayCR),
		reconcileOryOathkeeperConfigConfigMap(ctx, k8sClient, *apiGatewayCR),
		reconcileOathkeeperHPA(ctx, k8sClient, *apiGatewayCR),
		reconcileOryOathkeeperServiceAccount(ctx, k8sClient, *apiGatewayCR),
		reconcileOryOathkeeperServices(ctx, k8sClient, *apiGatewayCR),
		reconcileOathkeeperDeployment(ctx, k8sClient, *apiGatewayCR),
		reconcileOathkeeperPdb(ctx, k8sClient, *apiGatewayCR),
	)
	if err != nil {
		return controllers.ErrorStatus(err, "Oathkeeper did not reconcile successfully", conditions.OathkeeperReconcileFailed.Condition())
	}

	return controllers.ReadyStatus(conditions.OathkeeperReconcileSucceeded.Condition())
}
