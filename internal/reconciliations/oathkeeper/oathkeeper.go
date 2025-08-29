package oathkeeper

import (
	"context"
	"errors"
	"time"

	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/conditions"
	"github.com/kyma-project/api-gateway/internal/reconciliations"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper/maester"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	apiRuleConfigMapName        = "api-gateway-config.operator.kyma-project.io"
	apiRuleConfigMapNamespace   = "kyma-system"
	enableAPIRuleV1ConfigMapKey = "enableDeprecatedV1beta1APIRule"
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
	configMap := &corev1.ConfigMap{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Namespace: apiRuleConfigMapNamespace,
		Name:      apiRuleConfigMapName,
	}, configMap)
	if err != nil || configMap.Data[enableAPIRuleV1ConfigMapKey] != "true" {
		ctrl.Log.Info("Oathkeeper reconciliation disabled")
		return DeleteOathkeeper(ctx, k8sClient)
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

func DeleteOathkeeper(ctx context.Context, k8sClient client.Client) controllers.Status {
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
		deleteCRD(ctx, k8sClient, crdName),
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
