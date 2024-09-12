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
	"fmt"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	v2alpha1Processing "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/validation/v2alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	defaultReconciliationPeriod   = 30 * time.Minute
	errorReconciliationPeriod     = 1 * time.Minute
	migrationReconciliationPeriod = 1 * time.Minute
	apiGatewayFinalizer           = "gateway.kyma-project.io/subresources"
)

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
	l := r.Log.WithValues("namespace", req.Namespace, "APIRule", req.Name)
	l.Info("Starting reconciliation")
	ctx = logr.NewContext(ctx, r.Log)

	defaultDomainName, err := default_domain.GetDefaultDomainFromKymaGateway(ctx, r.Client)
	if err != nil && default_domain.HandleDefaultDomainError(l, err) {
		return doneReconcileErrorRequeue(err, errorReconciliationPeriod)
	}

	isCMReconcile := req.NamespacedName.String() == types.NamespacedName{
		Namespace: helpers.CM_NS, Name: helpers.CM_NAME}.String()

	finishReconcile := r.reconcileConfigMap(ctx, isCMReconcile)
	if finishReconcile {
		return doneReconcileNoRequeue()
	}

	apiRule := gatewayv1beta1.APIRule{}

	if err := r.Client.Get(ctx, req.NamespacedName, &apiRule); err != nil {
		if apierrs.IsNotFound(err) {
			return doneReconcileNoRequeue()
		}
		l.Error(err, "Error while getting APIRule")
		return doneReconcileErrorRequeue(err, errorReconciliationPeriod)
	}
	// assign LastProcessedTime and ObservedGeneration early to indicate that
	// resource got reconciled
	apiRule.Status.LastProcessedTime = metav1.Now()
	apiRule.Status.ObservedGeneration = apiRule.Generation

	if !apiRule.DeletionTimestamp.IsZero() {
		l.Info("APIRule is marked for deletion, deleting")
		return r.reconcileAPIRuleDeletion(ctx, l, &apiRule)
	}

	if !controllerutil.ContainsFinalizer(&apiRule, apiGatewayFinalizer) {
		l.Info("APIRule is missing a finalizer, adding")
		n := apiRule.DeepCopy()
		controllerutil.AddFinalizer(n, apiGatewayFinalizer)
		return r.updateResourceRequeue(ctx, l, n)
	}

	if r.isApiRuleConvertedFromV2alpha1(apiRule) {
		return r.reconcileV2alpha1APIRule(ctx, l, apiRule, defaultDomainName)
	}

	l.Info("Reconciling v1beta1 APIRule", "jwtHandler", r.Config.JWTHandler)
	cmd := r.getV1beta1Reconciliation(&apiRule, defaultDomainName, &l)
	if name, err := dependencies.APIRule().AreAvailable(ctx, r.Client); err != nil {
		s, err := handleDependenciesError(name, err).V1beta1Status()
		if err != nil {
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		if err := s.UpdateStatus(&apiRule.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.updateStatus(ctx, l, &apiRule)
	}

	l.Info("Validating APIRule config")
	failures := validation.ValidateConfig(r.Config)
	if len(failures) > 0 {
		l.Error(fmt.Errorf("validation has failures"),
			"Configuration validation failed", "failures", failures)
		s := cmd.GetStatusBase(string(gatewayv1beta1.StatusSkipped)).
			GenerateStatusFromFailures(failures)
		if err := s.UpdateStatus(&apiRule.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.updateStatus(ctx, l, &apiRule)
	}

	l.Info("Reconciling APIRule sub-resources")
	s := processing.Reconcile(ctx, r.Client, &l, cmd)
	if err := s.UpdateStatus(&apiRule.Status); err != nil {
		l.Error(err, "Error updating APIRule status")
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	return r.updateStatus(ctx, l, &apiRule)
}

func (r *APIRuleReconciler) reconcileV2alpha1APIRule(ctx context.Context, l logr.Logger, apiRule gatewayv1beta1.APIRule, domain string) (ctrl.Result, error) {
	l.Info("Reconciling v2alpha1 APIRule")
	toUpdate := apiRule.DeepCopy()
	migrate, err := apiRuleNeedsMigration(ctx, r.Client, toUpdate)
	if err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	if migrate {
		migration.ApplyMigrationAnnotation(l, toUpdate)
		// should not conflict with future status updates as long as there are
		// no Update() calls to the resource after that
		if err := r.Update(ctx, toUpdate); err != nil {
			l.Error(err, "Failed to update migration annotation")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
	}

	// convert v1beta1 to v2alpha1
	rule := gatewayv2alpha1.APIRule{}
	if err := rule.ConvertFrom(toUpdate); err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	cmd := r.getv2alpha1Reconciliation(&apiRule, &rule,
		domain, migrate, &l)

	if name, err := dependencies.APIRule().AreAvailable(ctx, r.Client); err != nil {
		s, err := handleDependenciesError(name, err).V2alpha1Status()
		if err != nil {
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		if err := s.UpdateStatus(&rule.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.convertAndUpdateStatus(ctx, l, rule)
	}

	l.Info("Validating APIRule config")
	failures := validation.ValidateConfig(r.Config)
	if len(failures) > 0 {
		l.Error(fmt.Errorf("validation has failures"),
			"Configuration validation failed", "failures", failures)
		s := cmd.GetStatusBase(string(gatewayv2alpha1.Error)).
			GenerateStatusFromFailures(failures)
		if err := s.UpdateStatus(&rule.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.convertAndUpdateStatus(ctx, l, rule)
	}

	l.Info("Reconciling APIRule sub-resources")
	s := processing.Reconcile(ctx, r.Client, &l, cmd)

	if err := s.UpdateStatus(&rule.Status); err != nil {
		l.Error(err, "Error updating APIRule status")
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	return r.convertAndUpdateStatus(ctx, l, rule)
}

// convertAndUpdateStatus is a small helper function that converts APIRule
// resource from convertible v2alpha1 to hub version v2alpha1
func (r *APIRuleReconciler) convertAndUpdateStatus(ctx context.Context, l logr.Logger,
	rule gatewayv2alpha1.APIRule) (ctrl.Result, error) {
	l.Info("Converting APIRule v2alpha1 to v1beta1")
	toUpdate := gatewayv1beta1.APIRule{}
	if err := rule.ConvertTo(&toUpdate); err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	return r.updateStatus(ctx, l, &toUpdate)
}

func (r *APIRuleReconciler) updateResourceRequeue(ctx context.Context,
	log logr.Logger, rule client.Object) (ctrl.Result, error) {
	log.Info("Updating APIRule Resource", "requeue", "true")
	if err := r.Update(ctx, rule); err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	return ctrl.Result{Requeue: true}, nil
}

func apiRuleNeedsMigration(ctx context.Context, k8sClient client.Client, apiRule *gatewayv1beta1.APIRule) (bool, error) {
	var ownedRules rulev1alpha1.RuleList
	labels := processing.GetOwnerLabels(apiRule)
	if err := k8sClient.List(ctx, &ownedRules, client.MatchingLabels(labels)); err != nil {
		return false, err
	}
	return len(ownedRules.Items) > 0, nil
}

func handleDependenciesError(name string, err error) controllers.Status {
	if apierrs.IsNotFound(err) {
		return controllers.WarningStatus(err, fmt.Sprintf("CRD %s is not present. Make sure to install required dependencies for the component", name), nil)
	} else {
		return controllers.ErrorStatus(err, "Error happened during discovering dependencies", nil)
	}
}

func (r *APIRuleReconciler) getV1beta1Reconciliation(apiRule *gatewayv1beta1.APIRule, defaultDomain string, namespacedLogger *logr.Logger) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	config.DefaultDomainName = defaultDomain
	switch {
	case r.Config.JWTHandler == helpers.JWT_HANDLER_ISTIO:
		return istio.NewIstioReconciliation(apiRule, config, namespacedLogger)
	default:
		return ory.NewOryReconciliation(apiRule, config, namespacedLogger)
	}
}

func (r *APIRuleReconciler) getv2alpha1Reconciliation(apiRulev1beta1 *gatewayv1beta1.APIRule, apiRulev2alpha1 *gatewayv2alpha1.APIRule, defaultDomain string, needsMigration bool, namespacedLogger *logr.Logger) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	config.DefaultDomainName = defaultDomain
	v2alpha1Validator := v2alpha1.NewAPIRuleValidator(apiRulev2alpha1)
	return v2alpha1Processing.NewReconciliation(apiRulev2alpha1, apiRulev1beta1, v2alpha1Validator, config, namespacedLogger, needsMigration)
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

func (r *APIRuleReconciler) isApiRuleConvertedFromV2alpha1(apiRule gatewayv1beta1.APIRule) bool {
	// If the ApiRule is not found, we don't need to do anything. If it's found and converted, CM reconciliation is not needed.
	if apiRule.Annotations != nil {
		if originalVersion, ok := apiRule.Annotations["gateway.kyma-project.io/original-version"]; ok && originalVersion == "v2alpha1" {
			r.Log.Info("ApiRule is converted from v2alpha1", "name", apiRule.Name, "namespace", apiRule.Namespace)
			return true
		}
	}
	return false
}
