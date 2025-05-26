/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package operator

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"

	"github.com/kyma-project/api-gateway/internal/conditions"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"

	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"

	"errors"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/reconciliations/gateway"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	APIGatewayResourceListDefaultPath = "manifests/controlled_resources_list.yaml"
	ApiGatewayFinalizer               = "gateways.operator.kyma-project.io/api-gateway"
	//defaultApiGatewayReconciliationInterval = time.Hour * 10
	// Temporarily reduced the interval to 1 hour to make sure that NLB migration does
	defaultApiGatewayReconciliationInterval = time.Hour
)

func NewAPIGatewayReconciler(mgr manager.Manager, oathkeeperReconciler ReadyVerifyingReconciler) *APIGatewayReconciler {
	return &APIGatewayReconciler{
		Client:               mgr.GetClient(),
		Scheme:               mgr.GetScheme(),
		log:                  mgr.GetLogger().WithName("apigateway-controller"),
		oathkeeperReconciler: oathkeeperReconciler,
	}
}

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets;deployments;services;serviceaccounts,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="oathkeeper.ory.sh",resources=rules,verbs=deletecollection;create;delete;get;list;patch;update;watch
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="cert.gardener.cloud",resources=certificates,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="dns.gardener.cloud",resources=dnsentries,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;update;patch;create;delete

func (r *APIGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log.Info("Received reconciliation request", "name", req.Name)

	apiGatewayCR := operatorv1alpha1.APIGateway{}
	if err := r.Get(ctx, req.NamespacedName, &apiGatewayCR); err != nil {
		if apierrors.IsNotFound(err) {
			r.log.Info("Skipped reconciliation, because ApiGateway CR was not found")
			return ctrl.Result{}, nil
		}
		r.log.Info("Could not get APIGateway CR")
		return ctrl.Result{}, err
	}

	existingAPIGateways := &operatorv1alpha1.APIGatewayList{}
	if err := r.List(ctx, existingAPIGateways); err != nil {
		r.log.Info("Unable to list APIGateway CRs")
		return r.requeueReconciliation(ctx, apiGatewayCR, controllers.ErrorStatus(err, "Unable to list APIGateway CRs", conditions.ReconcileFailed.Condition()))
	}

	if len(existingAPIGateways.Items) > 1 {
		oldestCr := operatorv1alpha1.GetOldestAPIGatewayCR(existingAPIGateways)
		if oldestCr == nil {
			err := fmt.Errorf("stopped APIGateway CR reconciliation: no oldest APIGateway CR found")
			return r.terminateReconciliation(ctx, apiGatewayCR, controllers.WarningStatus(err, err.Error(), conditions.ReconcileFailed.Condition()))
		}

		if apiGatewayCR.GetUID() != oldestCr.GetUID() {
			err := fmt.Errorf("stopped APIGateway CR reconciliation: only APIGateway CR %s reconciles the module", oldestCr.GetName())
			return r.terminateReconciliation(ctx, apiGatewayCR, controllers.WarningStatus(err, err.Error(), conditions.OlderCRExists.Condition()))
		}
	}

	r.log.Info("Reconciling APIGateway CR", "name", apiGatewayCR.Name, "isInDeletion", apiGatewayCR.IsInDeletion())

	if err := controllers.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, controllers.ProcessingStatus(conditions.ReconcileProcessing.Condition())); err != nil {
		r.log.Error(err, "Update status to processing failed")
		// We don't update the status to error, because the status update already failed and to avoid another status update error we simply requeue the request.
		return ctrl.Result{}, err
	}

	if !apiGatewayCR.IsInDeletion() {
		if name, dependenciesErr := dependencies.ApiGateway().AreAvailable(ctx, r.Client); dependenciesErr != nil {
			return r.requeueReconciliation(ctx, apiGatewayCR, handleDependenciesError(name, dependenciesErr))
		}
	}

	if finalizerStatus := r.reconcileFinalizer(ctx, &apiGatewayCR); !finalizerStatus.IsReady() {
		return r.requeueReconciliation(ctx, apiGatewayCR, finalizerStatus)
	}

	if kymaGatewayStatus := gateway.ReconcileKymaGateway(ctx, r.Client, &apiGatewayCR, APIGatewayResourceListDefaultPath); !kymaGatewayStatus.IsReady() {
		return r.requeueReconciliation(ctx, apiGatewayCR, kymaGatewayStatus)
	}

	if oryOathkeeperStatus := r.oathkeeperReconciler.ReconcileAndVerifyReadiness(ctx, r.Client, &apiGatewayCR); !oryOathkeeperStatus.IsReady() {
		return r.requeueReconciliation(ctx, apiGatewayCR, oryOathkeeperStatus)
	}

	// If there are no finalizers left, we must assume that the resource is deleted and therefore must stop the reconciliation
	// to prevent accidental read or write to the resource.
	if !apiGatewayCR.HasFinalizer() {
		r.log.Info("End reconciliation because all finalizers have been removed")
		return ctrl.Result{}, nil
	}

	return r.finishReconcile(ctx, apiGatewayCR)
}

func handleDependenciesError(name string, err error) controllers.Status {
	if apierrors.IsNotFound(err) {
		return controllers.ErrorStatus(err, fmt.Sprintf("CRD %s is not present. Make sure to install required dependencies for the component", name), conditions.DependenciesMissing.Condition())
	} else {
		return controllers.ErrorStatus(err, "Error happened during discovering dependencies", conditions.ReconcileFailed.Condition())
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIGatewayReconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.APIGateway{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			if obj.GetNamespace() != "istio-system" {
				return nil
			}

			if obj.GetName() != "istio-ingressgateway" {
				return nil
			}

			apiGatewayList := &operatorv1alpha1.APIGatewayList{}
			if err := r.Client.List(ctx, apiGatewayList); err != nil {
				return nil
			}

			apiGateway := operatorv1alpha1.GetOldestAPIGatewayCR(apiGatewayList)
			if apiGateway == nil {
				return nil
			}

			return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: apiGateway.Namespace, Name: apiGateway.Name}}}
		})).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(c),
		}).
		Complete(r)
}

// requeueReconciliation cancels the reconciliation and requeues the request.
func (r *APIGatewayReconciler) requeueReconciliation(ctx context.Context, cr operatorv1alpha1.APIGateway, status controllers.Status) (ctrl.Result, error) {
	r.log.Error(status.NestedError(), "Reconcile failed")

	statusUpdateErr := controllers.UpdateApiGatewayStatus(ctx, r.Client, &cr, status)
	if statusUpdateErr != nil {
		r.log.Error(statusUpdateErr, "Update status failed")
	}

	return ctrl.Result{}, status.NestedError()
}

func (r *APIGatewayReconciler) finishReconcile(ctx context.Context, cr operatorv1alpha1.APIGateway) (ctrl.Result, error) {
	if err := controllers.UpdateApiGatewayStatus(ctx, r.Client, &cr, controllers.ReadyStatus(conditions.ReconcileSucceeded.Condition())); err != nil {
		r.log.Error(err, "Update status failed")
		return ctrl.Result{}, err
	}

	r.log.Info("Successfully reconciled")
	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: defaultApiGatewayReconciliationInterval,
	}, nil
}

func (r *APIGatewayReconciler) terminateReconciliation(ctx context.Context, apiGatewayCR operatorv1alpha1.APIGateway, status controllers.Status) (ctrl.Result, error) {
	statusUpdateErr := controllers.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, status)

	if statusUpdateErr != nil {
		r.log.Error(statusUpdateErr, "Error during updating status to error")
		// In case the update of the status fails we must requeue the request, because otherwise the Error state is never visible in the CR.
		return ctrl.Result{}, statusUpdateErr
	}

	r.log.Error(status.NestedError(), "Reconcile failed, but won't requeue")
	return ctrl.Result{}, nil
}

func (r *APIGatewayReconciler) reconcileFinalizer(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) controllers.Status {
	if !apiGatewayCR.IsInDeletion() && !hasFinalizer(apiGatewayCR) {
		controllerutil.AddFinalizer(apiGatewayCR, ApiGatewayFinalizer)
		if err := r.Update(ctx, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Failed to add API-Gateway CR finalizer")
			return controllers.ErrorStatus(err, "Could not add API-Gateway CR finalizer", conditions.ReconcileFailed.Condition())
		}
	}

	if apiGatewayCR.IsInDeletion() && hasFinalizer(apiGatewayCR) {
		apiRulesFound, err := apiRulesExist(ctx, r.Client)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during listing existing APIRules", conditions.ReconcileFailed.Condition())
		}
		if len(apiRulesFound) > 0 {
			return controllers.WarningStatus(errors.New("could not delete API-Gateway CR since there are APIRule(s) that block its deletion"),
				"There are APIRule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				conditions.DeletionBlockedExistingResources.AdditionalMessage(": "+strings.Join(apiRulesFound, ", ")).Condition())
		}

		oryRulesFound, err := oryRulesExist(ctx, r.Client)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during listing existing ORY Oathkeeper Rules", conditions.ReconcileFailed.Condition())
		}
		if len(oryRulesFound) > 0 {
			return controllers.WarningStatus(errors.New("could not delete API-Gateway CR since there are ORY Oathkeeper Rule(s) that block its deletion"),
				"There are ORY Oathkeeper Rule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				conditions.DeletionBlockedExistingResources.AdditionalMessage(": "+strings.Join(oryRulesFound, ", ")).Condition())
		}
		rateLimiterRules, err := rateLimitsExists(ctx, r.Client)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during listing existing Rate Limit", conditions.ReconcileFailed.Condition())
		}
		if len(rateLimiterRules) > 0 {
			return controllers.WarningStatus(errors.New("could not delete API-Gateway CR since there are RateLimit(s) that block its deletion"),
				"There are RateLimit(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				conditions.DeletionBlockedExistingResources.AdditionalMessage(": "+strings.Join(rateLimiterRules, ", ")).Condition())
		}
		if err := removeFinalizer(ctx, r.Client, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Error happened during API-Gateway CR finalizer removal")
			return controllers.ErrorStatus(err, "Could not remove finalizer", conditions.ReconcileFailed.Condition())
		}
	}

	return controllers.ReadyStatus(conditions.ReconcileSucceeded.Condition())
}

func rateLimitsExists(ctx context.Context, k8sClient client.Client) ([]string, error) {
	rateLimits := ratelimitv1alpha1.RateLimitList{}
	err := k8sClient.List(ctx, &rateLimits)
	if meta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
		// RateLimits does not exist, there is not blocking rate limits
		return nil, nil
	}

	ctrl.Log.Info(fmt.Sprintf("There are %d RateLimit(s) found on cluster", len(rateLimits.Items)))
	var blockingRateLimits []string
	for _, rateLimit := range rateLimits.Items {
		rateLimitNamespacedName := rateLimit.GetNamespace() + "/" + rateLimit.GetName()
		ctrl.Log.Info("RateLimit rule blocking deletion", "rule", rateLimitNamespacedName)
		blockingRateLimits = append(blockingRateLimits, rateLimitNamespacedName)
	}
	return blockingRateLimits, nil
}

func apiRulesExist(ctx context.Context, k8sClient client.Client) ([]string, error) {
	apiRuleList := gatewayv2alpha1.APIRuleList{}
	err := k8sClient.List(ctx, &apiRuleList)
	if meta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
		// ApiRule CRD does not exist, there are no blocking rules
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ctrl.Log.Info(fmt.Sprintf("There are %d APIRule(s) found on cluster", len(apiRuleList.Items)))
	var blockingApiRules []string
	for _, rule := range apiRuleList.Items {
		blocking := rule.GetNamespace() + "/" + rule.GetName()
		ctrl.Log.Info("APIRule blocking deletion", "rule", blocking)
		if len(blockingApiRules) < 5 {
			blockingApiRules = append(blockingApiRules, blocking)
		}
	}
	return blockingApiRules, nil
}

func oryRulesExist(ctx context.Context, k8sClient client.Client) ([]string, error) {
	oryRulesList := oryv1alpha1.RuleList{}
	err := k8sClient.List(ctx, &oryRulesList)
	if meta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
		// Oathkeeper CRD does not exist, there are no blocking rules
		return nil, nil
	}
	if err != nil {
		// any other error
		return nil, err
	}
	ctrl.Log.Info(fmt.Sprintf("There are %d ORY Oathkeeper Rule(s) found on cluster", len(oryRulesList.Items)))
	var blockingOryRules []string
	for _, rule := range oryRulesList.Items {
		blocking := rule.GetNamespace() + "/" + rule.GetName()
		ctrl.Log.Info("ORY Oathkeeper rule blocking deletion", "rule", blocking)
		if len(blockingOryRules) < 5 {
			blockingOryRules = append(blockingOryRules, blocking)
		}
	}
	return blockingOryRules, nil
}

func hasFinalizer(apiGatewayCR *operatorv1alpha1.APIGateway) bool {
	return controllerutil.ContainsFinalizer(apiGatewayCR, ApiGatewayFinalizer)
}

func removeFinalizer(ctx context.Context, apiClient client.Client, apiGatewayCR *operatorv1alpha1.APIGateway) error {
	ctrl.Log.Info("Removing API-Gateway CR finalizer")
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := apiClient.Get(ctx, client.ObjectKeyFromObject(apiGatewayCR), apiGatewayCR); getErr != nil {
			return getErr
		}

		controllerutil.RemoveFinalizer(apiGatewayCR, ApiGatewayFinalizer)
		if updateErr := apiClient.Update(ctx, apiGatewayCR); updateErr != nil {
			return updateErr
		}

		ctrl.Log.Info("Successfully removed API-Gateway CR finalizer")
		return nil
	})
}
