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
	"regexp"
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/kyma-project/api-gateway/internal/gatewaytranslator"
	"github.com/kyma-project/api-gateway/internal/subresources/accessrule"

	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"

	"sigs.k8s.io/controller-runtime/pkg/event"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/processing/default_domain"
	"github.com/kyma-project/api-gateway/internal/processing/processors/istio"
	v2alpha1Processing "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	"github.com/kyma-project/api-gateway/internal/validation/v2alpha1"

	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing/processors/ory"
	"github.com/kyma-project/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/kyma-project/api-gateway/internal/processing"
)

const (
	defaultReconciliationPeriod   = 30 * time.Minute
	errorReconciliationPeriod     = 1 * time.Minute
	migrationReconciliationPeriod = 1 * time.Minute
	updateReconciliationPeriod    = 5 * time.Second
	apiGatewayFinalizer           = "gateway.kyma-project.io/subresources"
	oldGatewayFormatAnnotationKey = "gateway.kyma-project.io/old-gateway-format"
)

// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/finalizers,verbs=update
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=get;list;watch
// +kubebuilder:rbac:groups=oathkeeper.ory.sh,resources=rules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.istio.io,resources=authorizationpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.istio.io,resources=requestauthentications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch

func (r *APIRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := r.Log.WithValues("namespace", req.Namespace, "APIRule", req.Name)
	l.Info("Starting reconciliation")
	ctx = logr.NewContext(ctx, r.Log)

	defaultDomainName, err := default_domain.GetDomainFromKymaGateway(ctx, r.Client)
	if err != nil && default_domain.HandleDefaultDomainError(l, err) {
		return doneReconcileErrorRequeue(err, errorReconciliationPeriod)
	}

	isCMReconcile := req.String() == types.NamespacedName{
		Namespace: helpers.CM_NS, Name: helpers.CM_NAME}.String()

	finishReconcile := r.reconcileConfigMap(ctx, isCMReconcile)
	if finishReconcile {
		return doneReconcileNoRequeue()
	}
	apiRuleV2alpha1 := &gatewayv2alpha1.APIRule{}

	if err := r.Get(ctx, req.NamespacedName, apiRuleV2alpha1); err != nil {
		if apierrs.IsNotFound(err) {
			return doneReconcileNoRequeue()
		}
		l.Error(err, "Error while getting APIRule v2alpha1")
		return doneReconcileErrorRequeue(err, errorReconciliationPeriod)
	}

	// assign LastProcessedTime early to indicate that resource got reconciled
	apiRuleV2alpha1.Status.LastProcessedTime = metav1.Now()

	apiRule := gatewayv1beta1.APIRule{}
	err = apiRule.ConvertFrom(apiRuleV2alpha1)
	if err != nil {
		l.Error(err, "Error while converting APIRule v2alpha1 to v1beta1")
		return doneReconcileErrorRequeue(err, errorReconciliationPeriod)
	}
	apiRule.Status.ObservedGeneration = apiRule.Generation

	if !apiRule.DeletionTimestamp.IsZero() {
		l.Info("APIRule is marked for deletion, deleting")
		return r.reconcileAPIRuleDeletion(ctx, l, &apiRule)
	}
	if isAPIRuleV2(apiRuleV2alpha1) {
		return r.reconcileV2Alpha1APIRule(ctx, l, apiRuleV2alpha1, apiRule)
	}

	if !controllerutil.ContainsFinalizer(&apiRule, apiGatewayFinalizer) {
		l.Info("APIRule is missing a finalizer, adding")
		n := apiRule.DeepCopy()
		controllerutil.AddFinalizer(n, apiGatewayFinalizer)
		return r.updateResourceRequeue(ctx, l, n)
	}

	l.Info("Reconciling v1beta1 APIRule", "jwtHandler", r.Config.JWTHandler)
	cmd := r.getV1Beta1Reconciliation(&apiRule, defaultDomainName, &l)
	if name, err := dependencies.APIRuleV1beta1().AreAvailable(ctx, r.Client); err != nil {
		s, err := handleDependenciesError(name, err).V1beta1Status()
		if err != nil {
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		if err := s.UpdateStatus(&apiRule.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.convertAndUpdateStatus(ctx, l, apiRule, s.HasError())
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
		return r.convertAndUpdateStatus(ctx, l, apiRule, s.HasError())
	}

	l.Info("Reconciling APIRule sub-resources")
	s := processing.Reconcile(ctx, r.Client, &l, cmd)
	if err := s.UpdateStatus(&apiRule.Status); err != nil {
		l.Error(err, "Error updating APIRule status")
		// Quick retry if the object has been modified
		if strings.Contains(err.Error(), "the object has been modified") {
			r.Metrics.IncreaseApiRuleObjectModifiedErrorsCounter()
			return doneReconcileErrorRequeue(err, 0)
		}
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	if gatewaytranslator.IsOldGatewayNameFormat(*apiRule.Spec.Gateway) {
		// Update the status via API v1beta1 to avoid issues with CRD validation
		if apiRule.Status.APIRuleStatus.Code == gatewayv1beta1.StatusOK {
			apiRule.Status.APIRuleStatus.Code = gatewayv1beta1.StatusWarning
			apiRule.Status.APIRuleStatus.Description = "Version v1beta1 of APIRule is" +
				" deprecated and will be removed in future releases. Use version v2 instead."
		} else {
			apiRule.Status.APIRuleStatus.Description = fmt.Sprintf("Version v1beta1 of APIRule is deprecated and will"+
				" be removed in future releases. "+
				"Use version v2 instead.\n\n%s", apiRule.Status.APIRuleStatus.Description)
		}
		return r.updateStatus(ctx, l, &apiRule, s.HasError())
	}
	return r.convertAndUpdateStatus(ctx, l, apiRule, s.HasError())
}

func isAPIRuleV2(apiRule *gatewayv2alpha1.APIRule) bool {
	if originalVersion, ok := apiRule.Annotations["gateway.kyma-project.io/original-version"]; ok {
		return originalVersion != "v1beta1"
	}
	return true
}

func (r *APIRuleReconciler) reconcileV2Alpha1APIRule(ctx context.Context, l logr.Logger, apiRule *gatewayv2alpha1.APIRule, apiRuleV1beta1 gatewayv1beta1.APIRule) (ctrl.Result, error) {
	l.Info("Reconciling v2alpha1 APIRule")

	toUpdate := apiRule.DeepCopy()
	if !controllerutil.ContainsFinalizer(apiRule, apiGatewayFinalizer) {
		l.Info("APIRule is missing a finalizer, adding")
		n := apiRule.DeepCopy()
		controllerutil.AddFinalizer(n, apiGatewayFinalizer)
		return r.updateResourceRequeue(ctx, l, n)
	}

	migrate, err := apiRuleNeedsMigration(ctx, r.Client, &apiRuleV1beta1)
	if err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}

	l.Info("APIRule v2 before gateway discover", "apirule", toUpdate)
	gateway, err := discoverGateway(r.Client, ctx, l, toUpdate)
	if err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}

	if gateway == nil {
		return r.updateStatus(ctx, l, toUpdate, true)
	}

	cmd := r.getV2Alpha1Reconciliation(&apiRuleV1beta1, toUpdate, gateway, migrate, &l)

	if name, err := dependencies.APIRuleV2().AreAvailable(ctx, r.Client); err != nil {
		s, err := handleDependenciesError(name, err).V2alpha1Status()
		if err != nil {
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		if err := s.UpdateStatus(&toUpdate.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.updateStatus(ctx, l, toUpdate, s.HasError())
	}

	l.Info("Validating APIRule config")
	failures := validation.ValidateConfig(r.Config)
	if len(failures) > 0 {
		l.Error(fmt.Errorf("validation has failures"),
			"Configuration validation failed", "failures", failures)
		s := cmd.GetStatusBase(string(gatewayv2alpha1.Error)).
			GenerateStatusFromFailures(failures)
		if err := s.UpdateStatus(&toUpdate.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.updateStatus(ctx, l, toUpdate, s.HasError())
	}

	l.Info("Reconciling APIRule sub-resources")
	s := processing.Reconcile(ctx, r.Client, &l, cmd)

	if migrate && !s.HasError() {
		migration.ApplyMigrationAnnotation(l, toUpdate)
		// should not conflict with future status updates as long as there are
		// no Update() calls to the resource after that
		if err := r.Update(ctx, toUpdate); err != nil {
			l.Error(err, "Failed to update migration annotation")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
	}

	if err := s.UpdateStatus(&toUpdate.Status); err != nil {
		l.Error(err, "Error updating APIRule status")
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	return r.updateStatus(ctx, l, toUpdate, s.HasError())
}

func (r *APIRuleReconciler) updateResourceRequeue(ctx context.Context,
	log logr.Logger, rule client.Object) (ctrl.Result, error) {
	log.Info("Updating APIRule Resource", "requeue", "true")
	if err := r.Update(ctx, rule); err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	return ctrl.Result{RequeueAfter: updateReconciliationPeriod}, nil
}

func apiRuleNeedsMigration(ctx context.Context, k8sClient client.Client, apiRule *gatewayv1beta1.APIRule) (bool, error) {
	var crdOryRules apiextensionsv1.CustomResourceDefinition
	crdName := "rules.oathkeeper.ory.sh"
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: crdName}, &crdOryRules); err != nil {
		if apierrs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	repository := accessrule.NewRepository(k8sClient)
	oryRules, err := repository.GetAll(ctx, apiRule)
	if err != nil {
		return false, err
	}
	return len(oryRules) > 0, nil
}

func handleDependenciesError(name string, err error) controllers.Status {
	if apierrs.IsNotFound(err) {
		return controllers.WarningStatus(err, fmt.Sprintf("CRD %s is not present. Make sure to install required dependencies for the component", name), nil)
	} else {
		return controllers.ErrorStatus(err, "Error happened during discovering dependencies", nil)
	}
}

func discoverGateway(client client.Client, ctx context.Context, l logr.Logger, rule *gatewayv2alpha1.APIRule) (*networkingv1beta1.Gateway, error) {
	if rule.Spec.Gateway == nil {
		return nil, fmt.Errorf("expected Gateway to be set")
	}

	// The regex pattern is the exact same as the one used in the APIRule validation
	// Unfortunately usage of a constant is not possible here, as the Kubebuilder CRD validation interface is comment based
	match, err := regexp.MatchString(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?/([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)$`, *rule.Spec.Gateway)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, fmt.Errorf("expected Gateway %s to be in the namespace/name format", *rule.Spec.Gateway)
	}

	gatewayName := strings.Split(*rule.Spec.Gateway, "/")
	gatewayNN := types.NamespacedName{
		Namespace: gatewayName[0],
		Name:      gatewayName[1],
	}
	var gateway networkingv1beta1.Gateway
	if err := client.Get(ctx, gatewayNN, &gateway); err != nil {
		v2Alpha1Status := status.ReconciliationV2alpha1Status{
			ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
				State: gatewayv2alpha1.Error,
			},
		}
		s := v2Alpha1Status.GenerateStatusFromFailures([]validation.Failure{
			{
				AttributePath: "spec.gateway",
				Message:       "Could not get specified Gateway",
			},
		})
		if err := s.UpdateStatus(&rule.Status); err != nil {
			l.Error(err, "Error updating APIRule status")
			return nil, err
		}
		return nil, nil
	}

	return &gateway, nil
}

func (r *APIRuleReconciler) getV1Beta1Reconciliation(apiRule *gatewayv1beta1.APIRule, defaultDomainName string, namespacedLogger *logr.Logger) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	config.DefaultDomainName = defaultDomainName
	switch r.Config.JWTHandler {
	case helpers.JWT_HANDLER_ISTIO:
		return istio.NewIstioReconciliation(apiRule, config, namespacedLogger, r.Client)
	default:
		return ory.NewOryReconciliation(apiRule, config, namespacedLogger, r.Client)
	}
}

func (r *APIRuleReconciler) getV2Alpha1Reconciliation(apiRulev1beta1 *gatewayv1beta1.APIRule, apiRulev2alpha1 *gatewayv2alpha1.APIRule, gateway *networkingv1beta1.Gateway, needsMigration bool, namespacedLogger *logr.Logger) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	v2alpha1Validator := v2alpha1.NewAPIRuleValidator(apiRulev2alpha1)
	return v2alpha1Processing.NewReconciliation(apiRulev2alpha1, apiRulev1beta1, gateway, v2alpha1Validator, config, namespacedLogger, needsMigration, r.Client)
}

type annotationChangedPredicate = annotationChangedTypedPredicate[client.Object]

type annotationChangedTypedPredicate[object metav1.Object] struct {
	predicate.TypedFuncs[object]
	annotation string
}

func (p annotationChangedTypedPredicate[object]) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		return false
	}
	if e.ObjectNew == nil {
		return false
	}

	originalVersionOld, oldOK := e.ObjectOld.GetAnnotations()[p.annotation]
	originalVersionNew, newOK := e.ObjectNew.GetAnnotations()[p.annotation]
	return oldOK != newOK || originalVersionOld != originalVersionNew
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		// We need to filter for generation changes, because we had an issue that on Azure clusters the APIRules were constantly reconciled.
		For(&gatewayv2alpha1.APIRule{}, builder.WithPredicates(
			predicate.Or(
				predicate.GenerationChangedPredicate{},
				annotationChangedPredicate{annotation: "gateway.kyma-project.io/original-version"},
				annotationChangedPredicate{annotation: "gateway.kyma-project.io/v1beta1-spec"},
			))).
		Watches(&corev1.ConfigMap{}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(&isApiGatewayConfigMapPredicate{Log: r.Log})).
		Watches(&corev1.Service{}, NewServiceInformer(r)).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(c),
		}).
		Complete(r)
}

// convertAndUpdateStatus is a small helper function that converts APIRule
// resource from convertible v1beta1 to hub version v2alpha1
func (r *APIRuleReconciler) convertAndUpdateStatus(ctx context.Context, l logr.Logger, rule gatewayv1beta1.APIRule, hasError bool) (ctrl.Result, error) {
	toUpdate := gatewayv2alpha1.APIRule{}
	if err := rule.ConvertTo(&toUpdate); err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}

	// If the APIRule is v1beta1 in Ready state, we set the status to Warning
	// to indicate that the APIRule v1beta1 is deprecated and should be migrated to v2.
	if toUpdate.Status.State == gatewayv2alpha1.Ready {
		toUpdate.Status.State = gatewayv2alpha1.Warning
		toUpdate.Status.Description = "Version v1beta1 of APIRule is" +
			" deprecated and will be removed in future releases. Use version v2 instead."
	} else {
		toUpdate.Status.Description = fmt.Sprintf("Version v1beta1 of APIRule is deprecated and will"+
			" be removed in future releases. "+
			"Use version v2 instead.\n\n%s", toUpdate.Status.Description)
	}
	return r.updateStatus(ctx, l, &toUpdate, hasError)
}
