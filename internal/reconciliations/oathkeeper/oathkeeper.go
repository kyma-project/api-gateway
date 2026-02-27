package oathkeeper

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-project/api-gateway/internal/access"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/conditions"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/maester"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	accessAllowed, err := access.ShouldAllowAccessToV1Beta1(ctx, k8sClient)
	if client.IgnoreNotFound(err) != nil {
		ctrl.Log.Error(err, "Failed to check access to APIRule v1beta1")
	}

	if !accessAllowed {
		ctrl.Log.Info("Oathkeeper reconciliation disabled")
		return DeleteOathkeeperIfNoRulesLeft(ctx, k8sClient)
	}

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

func DeleteOathkeeperIfNoRulesLeft(ctx context.Context, k8sClient client.Client) controllers.Status {
	oryRules := &unstructured.UnstructuredList{}
	oryRules.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "oathkeeper.ory.sh",
		Version: "v1alpha1",
		Kind:    "Rule",
	})

	if err := k8sClient.List(ctx, oryRules); err != nil {
		if meta.IsNoMatchError(err) {
			return controllers.ReadyStatus(conditions.OathkeeperReconcileDisabled.Condition())
		}
		if !k8serrors.IsNotFound(err) {
			return controllers.ErrorStatus(err, "Failed to list Ory rules", conditions.OathkeeperReconcileFailed.Condition())
		}
	} else {
		if len(oryRules.Items) > 0 {
			return controllers.ReadyStatus(conditions.OathkeeperReconcileSucceeded.Condition())
		}
	}

	err := errors.Join(
		maester.DeleteMaester(ctx, k8sClient),
		deleteSecret(ctx, k8sClient, secretName, reconciliations.Namespace),
		deleteConfigmap(ctx, k8sClient, configMapName, reconciliations.Namespace),
		deleteHPA(ctx, k8sClient, hpaName),
		deleteServiceAccount(ctx, k8sClient, serviceAccountName, reconciliations.Namespace),
		deleteOathkeeperServices(ctx, k8sClient),
		deleteDeployment(ctx, k8sClient, deploymentName),
		deletePdb(ctx, k8sClient, pdbName, reconciliations.Namespace),
	)

	if err != nil {
		return controllers.ErrorStatus(err, "Oathkeeper did not delete properly", conditions.OathkeeperReconcileFailed.Condition())
	}

	return controllers.ReadyStatus(conditions.OathkeeperReconcileDisabled.Condition())
}
