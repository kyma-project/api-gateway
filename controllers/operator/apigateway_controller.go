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

	oryv1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"

	"errors"

	"github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/reconciliations/gateway"
	"github.com/kyma-project/api-gateway/internal/reconciliations/oathkeeper"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	namespace                         = "kyma-system"
	APIGatewayResourceListDefaultPath = "manifests/controlled_resources_list.yaml"
	ApiGatewayFinalizer               = "gateways.operator.kyma-project.io/api-gateway"
)

func NewAPIGatewayReconciler(mgr manager.Manager) *APIGatewayReconciler {
	return &APIGatewayReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		log:    mgr.GetLogger().WithName("apigateway-controller"),
	}
}

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=peerauthentications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets;deployments;services;serviceaccounts,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="oathkeeper.ory.sh",resources=rules,verbs=*
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="autoscaling",resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="cert.gardener.cloud",resources=certificates,verbs=get;list;watch;update;patch;create;delete
//+kubebuilder:rbac:groups="dns.gardener.cloud",resources=dnsentries,verbs=get;list;watch;update;patch;create;delete

func (r *APIGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log.Info("Received reconciliation request", "name", req.Name)

	apiGatewayCR := operatorv1alpha1.APIGateway{}
	if err := r.Client.Get(ctx, req.NamespacedName, &apiGatewayCR); err != nil {
		if apierrors.IsNotFound(err) {
			r.log.Info("Skipped reconciliation, because ApiGateway CR was not found")
			return ctrl.Result{}, nil
		}
		r.log.Info("Could not get APIGateway CR")
		return ctrl.Result{}, err
	}

	r.log.Info("Reconciling APIGateway CR", "name", apiGatewayCR.Name, "isInDeletion", apiGatewayCR.IsInDeletion())

	if err := controllers.UpdateApiGatewayStatus(ctx, r.Client, &apiGatewayCR, controllers.ProcessingStatus()); err != nil {
		r.log.Error(err, "Update status to processing failed")
		// We don't update the status to error, because the status update already failed and to avoid another status update error we simply requeue the request.
		return ctrl.Result{}, err
	}

	if finalizerStatus := r.reconcileFinalizer(ctx, &apiGatewayCR); !finalizerStatus.IsReady() {
		return r.requeueReconciliation(ctx, apiGatewayCR, finalizerStatus)
	}

	if kymaGatewayStatus := gateway.ReconcileKymaGateway(ctx, r.Client, &apiGatewayCR, APIGatewayResourceListDefaultPath); !kymaGatewayStatus.IsReady() {
		return r.requeueReconciliation(ctx, apiGatewayCR, kymaGatewayStatus)
	}

	if oryOathkeeperStatus := oathkeeper.ReconcileOathkeeper(ctx, r.Client, &apiGatewayCR); !oryOathkeeperStatus.IsReady() {
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

// SetupWithManager sets up the controller with the Manager.
func (r *APIGatewayReconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.APIGateway{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
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

func (r *APIGatewayReconciler) finishReconcile(ctx context.Context, cr v1alpha1.APIGateway) (ctrl.Result, error) {
	if err := controllers.UpdateApiGatewayStatus(ctx, r.Client, &cr, controllers.ReadyStatus()); err != nil {
		r.log.Error(err, "Update status failed")
		return ctrl.Result{}, err
	}

	r.log.Info("Successfully reconciled")
	return ctrl.Result{}, nil
}

func (i *APIGatewayReconciler) reconcileFinalizer(ctx context.Context, apiGatewayCR *operatorv1alpha1.APIGateway) controllers.Status {
	if !apiGatewayCR.IsInDeletion() && !hasFinalizer(apiGatewayCR) {
		controllerutil.AddFinalizer(apiGatewayCR, ApiGatewayFinalizer)
		if err := i.Client.Update(ctx, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Failed to add API-Gateway CR finalizer")
			return controllers.ErrorStatus(err, "Could not add API-Gateway CR finalizer")
		}
	}

	if apiGatewayCR.IsInDeletion() && hasFinalizer(apiGatewayCR) {
		apiRulesFound, err := apiRulesExist(ctx, i.Client)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during listing existing APIRules")
		}

		oryRulesFound, err := oryRulesExist(ctx, i.Client)
		if err != nil {
			return controllers.ErrorStatus(err, "Error during listing existing ORY Oathkeeper Rules")
		}

		if apiRulesFound || oryRulesFound {
			return controllers.WarningStatus(errors.New("could not delete API-Gateway CR since there are custom resources that block its deletion"),
				"There are custom resources that block the deletion of API-Gateway CR. Please take a look at kyma-system/api-gateway-controller-manager logs to see more information about the warning")
		}

		if err := removeFinalizer(ctx, i.Client, apiGatewayCR); err != nil {
			ctrl.Log.Error(err, "Error happened during API-Gateway CR finalizer removal")
			return controllers.ErrorStatus(err, "Could not remove finalizer")
		}
	}

	return controllers.ReadyStatus()
}

func apiRulesExist(ctx context.Context, k8sClient client.Client) (bool, error) {
	apiRuleList := v1beta1.APIRuleList{}
	err := k8sClient.List(ctx, &apiRuleList)
	if err != nil {
		return false, err
	}
	ctrl.Log.Info(fmt.Sprintf("Blocking deletion because %d APIRules found on cluster", len(apiRuleList.Items)))
	return len(apiRuleList.Items) > 0, nil
}

func oryRulesExist(ctx context.Context, k8sClient client.Client) (bool, error) {
	oryRulesList := oryv1alpha1.RuleList{}
	err := k8sClient.List(ctx, &oryRulesList)
	if err != nil {
		return false, err
	}
	ctrl.Log.Info(fmt.Sprintf("Blocking deletion because %d ORY Oathkeeper Rules found on cluster", len(oryRulesList.Items)))
	return len(oryRulesList.Items) > 0, nil
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
