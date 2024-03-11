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

package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing/istio"
	"github.com/kyma-project/api-gateway/internal/processing/ory"
	"github.com/kyma-project/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	CONFIGMAP_NAME                = "api-gateway-config"
	CONFIGMAP_NS                  = "kyma-system"
	DEFAULT_RECONCILIATION_PERIOD = 30 * time.Minute
	ERROR_RECONCILIATION_PERIOD   = time.Minute
	API_GATEWAY_FINALIZER         = "gateway.kyma-project.io/subresources"
)

type isApiGatewayConfigMapPredicate struct {
	Log logr.Logger
	predicate.Funcs
}

func (p isApiGatewayConfigMapPredicate) Create(e event.CreateEvent) bool {
	return p.Generic(event.GenericEvent(e))
}

func (p isApiGatewayConfigMapPredicate) Delete(e event.DeleteEvent) bool {
	return p.Generic(event.GenericEvent{
		Object: e.Object,
	})
}

func (p isApiGatewayConfigMapPredicate) Update(e event.UpdateEvent) bool {
	return p.Generic(event.GenericEvent{
		Object: e.ObjectNew,
	})
}

func (p isApiGatewayConfigMapPredicate) Generic(e event.GenericEvent) bool {
	if e.Object == nil {
		p.Log.Error(nil, "Generic event has no object", "event", e)
		return false
	}
	configMap, okCM := e.Object.(*corev1.ConfigMap)
	return okCM && configMap.GetNamespace() == CONFIGMAP_NS && configMap.GetName() == CONFIGMAP_NAME
}

//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch
//+kubebuilder:rbac:groups=oathkeeper.ory.sh,resources=rules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=authorizationpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=requestauthentications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch

func (r *APIRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Starting reconciliation", "namespacedName", req.NamespacedName.String())
	ctx = logr.NewContext(ctx, r.Log)

	defaultDomainName, err := default_domain.GetDefaultDomainFromKymaGateway(ctx, r.Client)
	if err != nil {
		if apierrs.IsNotFound(err) {
			r.Log.Error(err, "Default domain wasn't found. APIRules will require full host")
		} else {
			r.Log.Error(err, "Error getting default domain")
			return doneReconcileErrorRequeue(ERROR_RECONCILIATION_PERIOD)
		}
	}

	validator := validation.APIRuleValidator{
		DefaultDomainName: defaultDomainName,
	}

	isCMReconcile := req.NamespacedName.String() == types.NamespacedName{Namespace: helpers.CM_NS, Name: helpers.CM_NAME}.String()
	if isCMReconcile || r.Config.JWTHandler == "" {
		r.Log.Info("Starting ConfigMap reconciliation")
		err := r.Config.ReadFromConfigMap(ctx, r.Client)
		if err != nil {
			if apierrs.IsNotFound(err) {
				r.Log.Info(fmt.Sprintf(`ConfigMap %s in namespace %s was not found {"controller": "Api"}, will use default config`, helpers.CM_NAME, helpers.CM_NS))
				r.Config.ResetToDefault()
			} else {
				r.Log.Error(err, fmt.Sprintf(`could not read ConfigMap %s in namespace %s {"controller": "Api"}`, helpers.CM_NAME, helpers.CM_NS))
				r.Config.Reset()
			}
		}
		if isCMReconcile {
			configValidationFailures := validator.ValidateConfig(r.Config)
			r.Log.Info("ConfigMap changed", "config", r.Config)
			if len(configValidationFailures) > 0 {
				failuresJson, _ := json.Marshal(configValidationFailures)
				r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "Api", "failures": %s}`, string(failuresJson)))
			}
			r.Log.Info("ConfigMap reconciliation finished")
			return doneReconcileNoRequeue()
		}
	}
	r.Log.Info("Starting ApiRule reconciliation", "jwtHandler", r.Config.JWTHandler)

	cmd := r.getReconciliation(defaultDomainName)

	apiRule := &gatewayv1beta1.APIRule{}
	err = r.Client.Get(ctx, req.NamespacedName, apiRule)
	if err != nil {
		if apierrs.IsNotFound(err) {
			//There is no APIRule. Nothing to process, dependent objects will be garbage-collected.
			r.Log.Info(fmt.Sprintf("Finishing reconciliation as ApiRule '%s' does not exist.", req.NamespacedName))
			return doneReconcileNoRequeue()
		}

		r.Log.Error(err, "Error getting ApiRule")

		statusBase := cmd.GetStatusBase(gatewayv1beta1.StatusSkipped)
		errorMap := map[processing.ResourceSelector][]error{processing.OnApiRule: {err}}
		status := processing.GetStatusForErrorMap(errorMap, statusBase)
		return r.updateStatusOrRetry(ctx, apiRule, status)
	}

	r.Log.Info("Reconciling ApiRule", "name", apiRule.Name, "namespace", apiRule.Namespace, "resource version", apiRule.ResourceVersion)

	if apiRule.DeletionTimestamp.IsZero() {
		if name, err := dependencies.APIRule().AreAvailable(ctx, r.Client); err != nil {
			status, err := handleDependenciesError(name, err).ToAPIRuleStatus()
			if err != nil {
				return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
			}
			return r.updateStatusOrRetry(ctx, apiRule, status)
		}

		if !controllerutil.ContainsFinalizer(apiRule, API_GATEWAY_FINALIZER) {
			controllerutil.AddFinalizer(apiRule, API_GATEWAY_FINALIZER)
			if err := r.Update(ctx, apiRule); err != nil {
				return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(apiRule, API_GATEWAY_FINALIZER) {
			// finalizer is present on APIRule, so all subresources need to be deleted
			if err := processing.DeleteAPIRuleSubresources(r.Client, ctx, *apiRule); err != nil {
				r.Log.Error(err, "Error happened during deletion of APIRule subresources")
				// if removing subresources ends in error, return with retry
				// so that it can be retried
				return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
			}

			// remove finalizer so the reconcilation can proceed
			controllerutil.RemoveFinalizer(apiRule, API_GATEWAY_FINALIZER)
			if err := r.Update(ctx, apiRule); err != nil {
				r.Log.Error(err, "Error happened during finalizer removal")
				return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
			}
		}
		return doneReconcileNoRequeue()
	}

	r.Log.Info("Validating ApiRule config")
	configValidationFailures := validator.ValidateConfig(r.Config)
	if len(configValidationFailures) > 0 {
		failuresJson, _ := json.Marshal(configValidationFailures)
		r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, apiRule.Namespace, apiRule.Name, string(failuresJson)))
		statusBase := cmd.GetStatusBase(gatewayv1beta1.StatusSkipped)
		return r.updateStatusOrRetry(ctx, apiRule, processing.GenerateStatusFromFailures(configValidationFailures, statusBase))
	}

	status := processing.Reconcile(ctx, r.Client, &r.Log, cmd, apiRule)
	return r.updateStatusOrRetry(ctx, apiRule, status)
}

func handleDependenciesError(name string, err error) controllers.Status {
	if apierrs.IsNotFound(err) {
		return controllers.WarningStatus(err, fmt.Sprintf("CRD %s is not present. Make sure to install required dependencies for the component", name), nil)
	} else {
		return controllers.ErrorStatus(err, "Error happened during discovering dependencies", nil)
	}
}

func (r *APIRuleReconciler) getReconciliation(defaultDomain string) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	config.DefaultDomainName = defaultDomain
	if r.Config.JWTHandler == helpers.JWT_HANDLER_ISTIO {
		return istio.NewIstioReconciliation(config, &r.Log)
	}
	return ory.NewOryReconciliation(config, &r.Log)

}

// SetupWithManager sets up the controller with the Manager.
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		// We need to filter for generation changes, because we had an issue that on Azure clusters the APIRules were constantly reconciled.
		For(&gatewayv1beta1.APIRule{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&corev1.ConfigMap{}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(&isApiGatewayConfigMapPredicate{Log: r.Log})).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(c),
		}).
		Complete(r)
}

// Updates api status. If there was an error during update, returns the error so that entire reconcile loop is retried. If there is no error, returns a "reconcile success" value.
func (r *APIRuleReconciler) updateStatusOrRetry(ctx context.Context, api *gatewayv1beta1.APIRule, status processing.ReconciliationStatus) (ctrl.Result, error) {
	_, updateStatusErr := r.updateStatus(ctx, api, status)
	if updateStatusErr != nil {
		r.Log.Error(updateStatusErr, "Error updating ApiRule status, retrying")
		return retryReconcile(updateStatusErr) //controller retries to set the correct status eventually.
	}

	// If error happened during reconciliation (e.g. VirtualService conflict) requeue for reconciliation earlier
	if status.HasError() {
		r.Log.Info("Requeue for reconciliation because the status has an error")
		return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
	}

	return doneReconcileDefaultRequeue(r.ReconcilePeriod, &r.Log)
}

func doneReconcileNoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func doneReconcileDefaultRequeue(reconcilerPeriod time.Duration, logger *logr.Logger) (ctrl.Result, error) {
	after := DEFAULT_RECONCILIATION_PERIOD
	if reconcilerPeriod != 0 {
		after = reconcilerPeriod
	}

	logger.Info("Finished reconciliation and requeue", "requeue period", after)
	return ctrl.Result{RequeueAfter: after}, nil
}

func doneReconcileErrorRequeue(errorReconcilerPeriod time.Duration) (ctrl.Result, error) {
	after := ERROR_RECONCILIATION_PERIOD
	if errorReconcilerPeriod != 0 {
		after = errorReconcilerPeriod
	}
	return ctrl.Result{RequeueAfter: after}, nil
}

func retryReconcile(err error) (ctrl.Result, error) {
	return reconcile.Result{Requeue: true}, err
}

func (r *APIRuleReconciler) updateStatus(ctx context.Context, api *gatewayv1beta1.APIRule, status processing.ReconciliationStatus) (*gatewayv1beta1.APIRule, error) {
	api, err := r.getLatestApiRule(ctx, api)
	if err != nil {
		return nil, err
	}

	api.Status.ObservedGeneration = api.Generation
	api.Status.LastProcessedTime = &v1.Time{Time: time.Now()}
	api.Status.APIRuleStatus = status.ApiRuleStatus
	api.Status.VirtualServiceStatus = status.VirtualServiceStatus
	api.Status.AccessRuleStatus = status.AccessRuleStatus
	api.Status.RequestAuthenticationStatus = status.RequestAuthenticationStatus
	api.Status.AuthorizationPolicyStatus = status.AuthorizationPolicyStatus

	r.Log.Info("Updating ApiRule status", "status", api.Status)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err = r.Client.Status().Update(ctx, api)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return api, nil
}

func (r *APIRuleReconciler) getLatestApiRule(ctx context.Context, api *gatewayv1beta1.APIRule) (*gatewayv1beta1.APIRule, error) {
	apiRule := &gatewayv1beta1.APIRule{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: api.Name, Namespace: api.Namespace}, apiRule)
	if err != nil {
		if apierrs.IsNotFound(err) {
			r.Log.Error(err, "ApiRule not found")
			return nil, err
		}

		r.Log.Error(err, "Error getting ApiRule")
		return nil, err
	}

	return apiRule, nil
}
