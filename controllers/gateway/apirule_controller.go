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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"time"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/controllers"
	"github.com/kyma-project/api-gateway/internal/dependencies"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	v2alpha1Processing "github.com/kyma-project/api-gateway/internal/processing/processors/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	"github.com/kyma-project/api-gateway/internal/validation/v2alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	defaultReconciliationPeriod   = 30 * time.Minute
	errorReconciliationPeriod     = 1 * time.Minute
	migrationReconciliationPeriod = 1 * time.Minute
	apiGatewayFinalizer           = "gateway.kyma-project.io/subresources"
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

	isCMReconcile := req.String() == types.NamespacedName{
		Namespace: helpers.CM_NS, Name: helpers.CM_NAME}.String()

	finishReconcile := r.reconcileConfigMap(ctx, isCMReconcile)
	if finishReconcile {
		return doneReconcileNoRequeue()
	}
	apiRule := &gatewayv2alpha1.APIRule{}

	if err := r.Get(ctx, req.NamespacedName, apiRule); err != nil {
		if apierrs.IsNotFound(err) {
			return doneReconcileNoRequeue()
		}
		l.Error(err, "Error while getting APIRule v2alpha1")
		return doneReconcileErrorRequeue(err, errorReconciliationPeriod)
	}

	// assign LastProcessedTime and ObservedGeneration early to indicate that
	// resource got reconciled
	apiRule.Status.LastProcessedTime = metav1.Now()

	if !apiRule.DeletionTimestamp.IsZero() {
		l.Info("APIRule is marked for deletion, deleting")
		return r.reconcileAPIRuleDeletion(ctx, l, apiRule)
	}

	return r.reconcileV2Alpha1APIRule(ctx, l, apiRule)
}

func isAPIRuleV2(apiRule *gatewayv2alpha1.APIRule) bool {
	if originalVersion, ok := apiRule.Annotations["gateway.kyma-project.io/original-version"]; ok {
		return originalVersion != "v1beta1"
	}
	return true
}

func (r *APIRuleReconciler) reconcileV2Alpha1APIRule(ctx context.Context, l logr.Logger, apiRule *gatewayv2alpha1.APIRule) (ctrl.Result, error) {
	l.Info("Reconciling v2alpha1 APIRule")

	toUpdate := apiRule.DeepCopy()
	l.Info("APIRule v2", "apirule", apiRule)
	if !controllerutil.ContainsFinalizer(apiRule, apiGatewayFinalizer) {
		l.Info("APIRule is missing a finalizer, adding")
		n := apiRule.DeepCopy()
		controllerutil.AddFinalizer(n, apiGatewayFinalizer)
		return r.updateResourceRequeue(ctx, l, n)
	}

	migrate, err := apiRuleNeedsMigration(ctx, r.Client, apiRule)
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

	l.Info("APIRule v2 before gateway discover", "apirule", toUpdate)
	gateway, err := discoverGateway(r.Client, ctx, l, toUpdate)
	if err != nil {
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}

	if gateway == nil {
		return r.updateStatus(ctx, l, toUpdate, true)
	}

	cmd := r.getV2Alpha1Reconciliation(apiRule, gateway, migrate, &l)

	if name, err := dependencies.APIRule().AreAvailable(ctx, r.Client); err != nil {
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
	return ctrl.Result{Requeue: true}, nil
}

func apiRuleNeedsMigration(ctx context.Context, k8sClient client.Client, apiRule *gatewayv2alpha1.APIRule) (bool, error) {
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

func (r *APIRuleReconciler) getV2Alpha1Reconciliation(apiRulev2alpha1 *gatewayv2alpha1.APIRule, gateway *networkingv1beta1.Gateway, needsMigration bool, namespacedLogger *logr.Logger) processing.ReconciliationCommand {
	config := r.ReconciliationConfig
	v2alpha1Validator := v2alpha1.NewAPIRuleValidator(apiRulev2alpha1)
	return v2alpha1Processing.NewReconciliation(apiRulev2alpha1, gateway, v2alpha1Validator, config, namespacedLogger, needsMigration)
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager, c controllers.RateLimiterConfig) error {
	return ctrl.NewControllerManagedBy(mgr).
		// We need to filter for generation changes, because we had an issue that on Azure clusters the APIRules were constantly reconciled.
		For(&gatewayv2alpha1.APIRule{}, builder.WithPredicates(
			predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
			))).
		Watches(&corev1.ConfigMap{}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(&isApiGatewayConfigMapPredicate{Log: r.Log})).
		Watches(&corev1.Service{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			var apiRules gatewayv2alpha1.APIRuleList
			if err := r.Client.List(ctx, &apiRules); err != nil {
				return nil
			}

			if len(apiRules.Items) == 0 {
				return nil
			}

			var requests []reconcile.Request

			for _, apiRule := range apiRules.Items {
				// match if service is exposed by an APIRule
				// and add APIRule to the reconciliation queue
				matches := func(target *gatewayv2alpha1.Service) bool {
					if target == nil {
						return false
					}

					matchesNs := apiRule.Namespace == obj.GetNamespace()
					if target.Namespace != nil {
						matchesNs = *target.Namespace == obj.GetNamespace()
					}

					var matchesName bool
					if target.Name != nil {
						matchesName = *target.Name == obj.GetName()
					}

					return matchesNs && matchesName
				}
				if matches(apiRule.Spec.Service) {
					requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
						Namespace: apiRule.Namespace,
						Name:      apiRule.Name,
					}})
					continue
				}
				for _, rule := range apiRule.Spec.Rules {
					if matches(rule.Service) {
						requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
							Namespace: apiRule.Namespace,
							Name:      apiRule.Name,
						}})
						continue
					}
				}
			}
			return requests
		})).
		WithOptions(controller.Options{
			RateLimiter: controllers.NewRateLimiter(c),
		}).
		Complete(r)
}

// convertAndUpdateStatus is a small helper function that converts APIRule
// resource from convertible v1beta1 to hub version v2alpha1
func (r *APIRuleReconciler) convertAndUpdateStatus(ctx context.Context, l logr.Logger, toUpdate gatewayv2alpha1.APIRule, hasError bool) (ctrl.Result, error) {
	return r.updateStatus(ctx, l, &toUpdate, hasError)
}
