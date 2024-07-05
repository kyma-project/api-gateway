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
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	v2alpha1Processing "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing/processors/ory"
	"github.com/kyma-project/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	defaultReconciliationPeriod = 30 * time.Minute
	errorReconciliationPeriod   = time.Minute
	apiGatewayFinalizer         = "gateway.kyma-project.io/subresources"
)

func (r *APIRuleReconciler) handleAPIRuleGetError(ctx context.Context, name types.NamespacedName, apiRule *gatewayv1beta1.APIRule, err error, cmd processing.ReconciliationCommand) (ctrl.Result, error) {
	if apierrs.IsNotFound(err) {
		//There is no APIRule. Nothing to process, dependent objects will be garbage-collected.
		r.Log.Info(fmt.Sprintf("Finishing reconciliation as ApiRule '%s' does not exist.", name))
		return doneReconcileNoRequeue()
	}

	r.Log.Error(err, "Error getting ApiRule")

	statusBase := cmd.GetStatusBase(string(gatewayv1beta1.StatusSkipped))
	errorMap := map[status.ResourceSelector][]error{status.OnApiRule: {err}}
	return r.updateStatusOrRetry(ctx, apiRule, statusBase.GetStatusForErrorMap(errorMap))
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
	if err != nil && default_domain.HandleDefaultDomainError(r.Log, err) {
		return doneReconcileErrorRequeue(errorReconciliationPeriod)
	}

	isCMReconcile := req.NamespacedName.String() == types.NamespacedName{Namespace: helpers.CM_NS, Name: helpers.CM_NAME}.String()

	finishReconcile := r.reconcileConfigMap(ctx, isCMReconcile)
	if finishReconcile {
		return doneReconcileNoRequeue()
	}

	r.Log.Info("Starting ApiRule reconciliation", "jwtHandler", r.Config.JWTHandler)

	apiRule := &gatewayv1beta1.APIRule{}
	apiRuleErr := r.Client.Get(ctx, req.NamespacedName, apiRule)
	var cmd processing.ReconciliationCommand
	if apiRuleErr == nil && r.isApiRuleConvertedFromv2alpha1(*apiRule) {
		apiRuleV2alpha1 := &gatewayv2alpha1.APIRule{}
		if err := r.Client.Get(ctx, req.NamespacedName, apiRuleV2alpha1); err != nil {
			return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
		}
		cmd = r.getv2alpha1Reconciliation(apiRule, apiRuleV2alpha1, defaultDomainName)
	} else {
		cmd = r.getv1beta1Reconciliation(apiRule, defaultDomainName)
	}

	if apiRuleErr != nil {
		return r.handleAPIRuleGetError(ctx, req.NamespacedName, apiRule, apiRuleErr, cmd)
	}

	r.Log.Info("Reconciling ApiRule", "name", apiRule.Name, "namespace", apiRule.Namespace, "resource version", apiRule.ResourceVersion)

	shouldDeleteAPIRule := !apiRule.DeletionTimestamp.IsZero()
	if !shouldDeleteAPIRule {
		if name, err := dependencies.APIRule().AreAvailable(ctx, r.Client); err != nil {
			status, err := handleDependenciesError(name, err).ToAPIRuleStatus()
			if err != nil {
				return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
			}
			return r.updateStatusOrRetry(ctx, apiRule, status)
		}

		if !controllerutil.ContainsFinalizer(apiRule, apiGatewayFinalizer) {
			controllerutil.AddFinalizer(apiRule, apiGatewayFinalizer)
			if err := r.Update(ctx, apiRule); err != nil {
				return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
			}
		}
	} else {
		return r.reconcileAPIRuleDeletion(ctx, apiRule)
	}

	r.Log.Info("Validating ApiRule config")
	configValidationFailures := validation.ValidateConfig(r.Config)
	if len(configValidationFailures) > 0 {
		failuresJson, _ := json.Marshal(configValidationFailures)
		r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "ApiRule", "request": "%s/%s", "failures": %s}`, apiRule.Namespace, apiRule.Name, string(failuresJson)))
		statusBase := cmd.GetStatusBase(string(gatewayv1beta1.StatusSkipped))
		return r.updateStatusOrRetry(ctx, apiRule, statusBase.GenerateStatusFromFailures(configValidationFailures))
	}

	s := processing.Reconcile(ctx, r.Client, &r.Log, cmd, req)
	return r.updateStatusOrRetry(ctx, apiRule, s)
}

func handleDependenciesError(name string, err error) controllers.Status {
	if apierrs.IsNotFound(err) {
		return controllers.WarningStatus(err, fmt.Sprintf("CRD %s is not present. Make sure to install required dependencies for the component", name), nil)
	} else {
		return controllers.ErrorStatus(err, "Error happened during discovering dependencies", nil)
	}
}

func (r *APIRuleReconciler) getv1beta1Reconciliation(apiRule *gatewayv1beta1.APIRule, defaultDomain string) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	config.DefaultDomainName = defaultDomain
	switch {
	case r.Config.JWTHandler == helpers.JWT_HANDLER_ISTIO:
		return istio.NewIstioReconciliation(apiRule, config, &r.Log)
	default:
		return ory.NewOryReconciliation(apiRule, config, &r.Log)
	}
}

func (r *APIRuleReconciler) getv2alpha1Reconciliation(apiRulev1beta1 *gatewayv1beta1.APIRule, apiRulev2alpha1 *gatewayv2alpha1.APIRule, defaultDomain string) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	config.DefaultDomainName = defaultDomain
	return v2alpha1Processing.NewReconciliation(apiRulev2alpha1, apiRulev1beta1, config, &r.Log)
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

func (r *APIRuleReconciler) isApiRuleConvertedFromv2alpha1(apiRule gatewayv1beta1.APIRule) bool {
	// If the ApiRule is not found, we don't need to do anything. If it's found and converted, CM reconciliation is not needed.
	if apiRule.Annotations != nil {
		if originalVersion, ok := apiRule.Annotations["gateway.kyma-project.io/original-version"]; ok && originalVersion == "v2alpha1" {
			r.Log.Info("ApiRule is converted from v2alpha1")
			return true
		}
	}

	return false
}
