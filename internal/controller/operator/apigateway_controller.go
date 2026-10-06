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
	"strings"
	"time"

	"github.com/kyma-project/api-gateway/internal/networkpolicy"
	"github.com/kyma-project/api-gateway/internal/vpa"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"

	oryv1alpha1 "github.com/kyma-project/api-gateway/internal/types/ory/oathkeeper-maester/api/v1alpha1"

	"errors"

	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/controller"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/reconciliations/gateway"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimecontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	APIGatewayResourceListDefaultPath = "manifests/controlled_resources_list.yaml"
	ApiGatewayFinalizer               = "gateways.operator.kyma-project.io/api-gateway"
	//defaultApiGatewayReconciliationInterval = time.Hour * 10
	// Temporarily reduced the interval to 1 hour to make sure that NLB migration does
	defaultApiGatewayReconciliationInterval = time.Hour
	certificateRequeueInterval              = time.Second * 15
)

func NewAPIGatewayReconciler(mgr manager.Manager, oathkeeperReconciler ReadyVerifyingReconciler) *APIGatewayReconciler {
	return &APIGatewayReconciler{
		Client:               mgr.GetClient(),
		Scheme:               mgr.GetScheme(),
		log:                  mgr.GetLogger().WithName("apigateway-controller"),
		oathkeeperReconciler: oathkeeperReconciler,
	}
}

// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;deployments;services;serviceaccounts,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups="oathkeeper.ory.sh",resources=rules,verbs=deletecollection;create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups="cert.gardener.cloud",resources=certificates,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups="dns.gardener.cloud",resources=dnsentries,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups="policy",resources=poddisruptionbudgets,verbs=get;list;watch;update;patch;create;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=create;deletecollection;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=autoscaling.k8s.io,resources=verticalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.k8s.io,resources=verticalpodautoscalercheckpoints,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch

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
		msg := "Unable to list APIGateway CRs"
		if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Error, msg, []metav1.Condition{operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, msg)}); statusErr != nil {
			r.log.Error(statusErr, "Update status failed")
		}
		return r.requeueReconciliation(err)
	}
	if len(existingAPIGateways.Items) > 1 {
		oldestCr := operatorv1alpha1.GetOldestAPIGatewayCR(existingAPIGateways)
		if oldestCr == nil {
			err := fmt.Errorf("stopped APIGateway CR reconciliation: no oldest APIGateway CR found")
			if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Warning, err.Error(), []metav1.Condition{operatorv1alpha1.WarningCondition(operatorv1alpha1.ReasonReconcileFailed, err.Error())}); statusErr != nil {
				r.log.Error(statusErr, "Update status failed")
			}
			return r.terminateReconciliation(err)
		}
		if apiGatewayCR.GetUID() != oldestCr.GetUID() {
			err := fmt.Errorf("stopped APIGateway CR reconciliation: only APIGateway CR %s reconciles the module", oldestCr.GetName())
			if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Warning, err.Error(), []metav1.Condition{operatorv1alpha1.WarningCondition(operatorv1alpha1.ReasonOlderCRExists, err.Error())}); statusErr != nil {
				r.log.Error(statusErr, "Update status failed")
			}
			return r.terminateReconciliation(err)
		}
	}

	r.log.Info("Reconciling APIGateway CR", "name", apiGatewayCR.Name, "isInDeletion", apiGatewayCR.IsInDeletion())

	networkPoliciesEnabled := apiGatewayCR.Spec.NetworkPoliciesEnabled != nil && *apiGatewayCR.Spec.NetworkPoliciesEnabled
	r.log.Info("Handling NetworkPolicies if needed", "networkPoliciesEnabled", networkPoliciesEnabled)
	opPolicy := networkpolicy.OperatorPolicy{
		Client:  r.Client,
		Enabled: networkPoliciesEnabled,
		Owner:   &apiGatewayCR,
	}
	if err := opPolicy.Handle(ctx); err != nil {
		if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Error, err.Error(), []metav1.Condition{operatorv1alpha1.NetworkPolicyErrorCondition(err.Error())}); statusErr != nil {
			r.log.Error(statusErr, "Update status failed")
		}
		return r.requeueReconciliation(err)
	}
	if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, apiGatewayCR.Status.State, apiGatewayCR.Status.Description, []metav1.Condition{operatorv1alpha1.NetworkPolicyReadyCondition()}); statusErr != nil {
		r.log.Error(statusErr, "Update status failed")
	}

	if r.shouldSetProcessing(ctx, req.NamespacedName) {
		if err := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Processing, "Reconciling APIGateway CR", []metav1.Condition{operatorv1alpha1.ProcessingCondition()}); err != nil {
			r.log.Error(err, "Update status to processing failed")
			return ctrl.Result{}, err
		}
	}

	if !apiGatewayCR.IsInDeletion() {
		if name, dependenciesErr := dependencies.ApiGateway().AreAvailable(ctx, r.Client); dependenciesErr != nil {
			readyCond, depCond := dependenciesErrorConditions(name, dependenciesErr)
			if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Error, readyCond.Message, []metav1.Condition{readyCond, depCond}); statusErr != nil {
				r.log.Error(statusErr, "Update status failed")
			}
			return ctrl.Result{}, dependenciesErr
		}
	}

	if state, desc, condition, err := r.reconcileFinalizer(ctx, &apiGatewayCR); err != nil {
		// reconcileFinalizer returns a valid condition for all error paths
		conditions := []metav1.Condition{condition}
		if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, state, desc, conditions); statusErr != nil {
			r.log.Error(statusErr, "Update status failed")
		}
		return r.requeueReconciliation(err)
	}

	if gwState, gwDesc, gwCond, gwErr := gateway.ReconcileKymaGateway(ctx, r.Client, &apiGatewayCR, APIGatewayResourceListDefaultPath); gwErr != nil || gwState == operatorv1alpha1.Processing {
		conditions := []metav1.Condition{gwCond}
		if gwState == operatorv1alpha1.Processing {
			if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Processing, gwDesc, conditions); statusErr != nil {
				return ctrl.Result{}, statusErr
			}
			return ctrl.Result{RequeueAfter: certificateRequeueInterval}, nil
		}
		if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, gwState, gwDesc, conditions); statusErr != nil {
			r.log.Error(statusErr, "Update status failed")
		}
		return r.requeueReconciliation(gwErr)
	}

	if oathkeeperState, oathkeeperDesc, oathkeeperCond, oaErr := r.oathkeeperReconciler.ReconcileAndVerifyReadiness(ctx, r.Client, &apiGatewayCR); oaErr != nil {
		if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, oathkeeperState, oathkeeperDesc, []metav1.Condition{oathkeeperCond}); statusErr != nil {
			r.log.Error(statusErr, "Update status failed")
		}
		return r.requeueReconciliation(oaErr)
	}

	r.log.Info("Reconciling VPA if CRD is available")
	vpaReconciler := vpa.NewReconciler(r.Client)
	if err := vpaReconciler.Reconcile(ctx, apiGatewayCR.IsInDeletion()); err != nil {
		msg := "Error during VPA reconciliation"
		if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Error, msg, []metav1.Condition{operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, msg)}); statusErr != nil {
			r.log.Error(statusErr, "Update status failed")
		}
		return r.requeueReconciliation(err)
	}

	// If there are no finalizers left, we must assume that the resource is deleted and therefore must stop the reconciliation
	// to prevent accidental read or write to the resource.
	if !apiGatewayCR.HasFinalizer() {
		r.log.Info("End reconciliation because all finalizers have been removed")
		return ctrl.Result{}, nil
	}

	if statusErr := controller.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, operatorv1alpha1.Ready, "Successfully reconciled", []metav1.Condition{operatorv1alpha1.ReadyCondition()}); statusErr != nil {
		r.log.Error(statusErr, "Update status failed")
		return ctrl.Result{}, statusErr
	}

	return r.finishReconcile()
}

func dependenciesErrorConditions(name string, err error) (readyCond, depCond metav1.Condition) {
	if apierrors.IsNotFound(err) {
		msg := fmt.Sprintf("CRD %s is not present. Make sure to install required dependencies for the component", name)
		return operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, msg),
			operatorv1alpha1.DependenciesMissingCondition(msg)
	}
	msg := "Error happened during discovering dependencies"
	return operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, msg),
		operatorv1alpha1.DependenciesErrorCondition(msg)
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIGatewayReconciler) SetupWithManager(mgr ctrl.Manager, c controller.RateLimiterConfig) error {
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
		Watches(&networkingv1.NetworkPolicy{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			if obj.GetNamespace() != "kyma-system" {
				// we only care about kyma-system
				return nil
			}
			labels := obj.GetLabels()
			if labels == nil {
				return nil
			}
			gatewayName, ok := labels[networkpolicy.OwningResourceLabel]
			if !ok {
				return nil
			}
			req := []reconcile.Request{{NamespacedName: types.NamespacedName{Name: gatewayName}}}
			return req
		})).
		WithOptions(runtimecontroller.Options{
			RateLimiter: controller.NewRateLimiter(c),
		}).
		Complete(r)
}

// requeueReconciliation requeues the request on reconciliation failure.
func (r *APIGatewayReconciler) requeueReconciliation(err error) (ctrl.Result, error) {
	r.log.Error(err, "Reconcile failed")
	return ctrl.Result{}, err
}

// finishReconcile returns success with requeue interval.
func (r *APIGatewayReconciler) finishReconcile() (ctrl.Result, error) {
	r.log.Info("Successfully reconciled")
	return ctrl.Result{RequeueAfter: defaultApiGatewayReconciliationInterval}, nil
}

// terminateReconciliation returns without requeue on terminal failure.
func (r *APIGatewayReconciler) terminateReconciliation(err error) (ctrl.Result, error) {
	r.log.Error(err, "Reconcile failed, but won't requeue")
	return ctrl.Result{}, nil
}

func (r *APIGatewayReconciler) reconcileFinalizer(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) (operatorv1alpha1.State, string, metav1.Condition, error) {
	if !apiGatewayCR.IsInDeletion() && !hasFinalizer(apiGatewayCR) {
		controllerutil.AddFinalizer(apiGatewayCR, ApiGatewayFinalizer)
		if err := r.Update(ctx, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Failed to add API-Gateway CR finalizer")
			return operatorv1alpha1.Error, "Could not add API-Gateway CR finalizer", operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, "Could not add API-Gateway CR finalizer"), err
		}
	}

	if apiGatewayCR.IsInDeletion() && hasFinalizer(apiGatewayCR) {
		apiRulesFound, err := apiRulesExist(ctx, r.Client)
		if err != nil {
			return operatorv1alpha1.Error, "Error during listing existing APIRules", operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, "Error during listing existing APIRules"), err
		}
		if len(apiRulesFound) > 0 {
			msg := "API Gateway deletion blocked because of the existing custom resources: " + strings.Join(apiRulesFound, ", ")
			return operatorv1alpha1.Warning, "There are APIRule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				operatorv1alpha1.WarningCondition(operatorv1alpha1.ReasonDeletionBlockedExistingResources, msg),
				errors.New("could not delete API-Gateway CR since there are APIRule(s) that block its deletion")
		}

		oryRulesFound, err := oryRulesExist(ctx, r.Client)
		if err != nil {
			return operatorv1alpha1.Error, "Error during listing existing ORY Oathkeeper Rules", operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, "Error during listing existing ORY Oathkeeper Rules"), err
		}
		if len(oryRulesFound) > 0 {
			msg := "API Gateway deletion blocked because of the existing custom resources: " + strings.Join(oryRulesFound, ", ")
			return operatorv1alpha1.Warning, "There are ORY Oathkeeper Rule(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				operatorv1alpha1.WarningCondition(operatorv1alpha1.ReasonDeletionBlockedExistingResources, msg),
				errors.New("could not delete API-Gateway CR since there are ORY Oathkeeper Rule(s) that block its deletion")
		}

		rateLimiterRules, err := rateLimitsExists(ctx, r.Client)
		if err != nil {
			return operatorv1alpha1.Error, "Error during listing existing Rate Limit", operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, "Error during listing existing Rate Limit"), err
		}
		if len(rateLimiterRules) > 0 {
			msg := "API Gateway deletion blocked because of the existing custom resources: " + strings.Join(rateLimiterRules, ", ")
			return operatorv1alpha1.Warning, "There are RateLimit(s) that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning",
				operatorv1alpha1.WarningCondition(operatorv1alpha1.ReasonDeletionBlockedExistingResources, msg),
				errors.New("could not delete API-Gateway CR since there are RateLimit(s) that block its deletion")
		}

		if err := removeFinalizer(ctx, r.Client, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Error happened during API-Gateway CR finalizer removal")
			return operatorv1alpha1.Error, "Could not remove finalizer", operatorv1alpha1.ErrorCondition(operatorv1alpha1.ReasonReconcileFailed, "Could not remove finalizer"), err
		}
	}

	return operatorv1alpha1.Ready, "", metav1.Condition{}, nil
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
	apiRuleList := v1beta1.APIRuleList{}
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
