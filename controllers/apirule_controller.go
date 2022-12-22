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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/processing/istio"
	"github.com/kyma-incubator/api-gateway/internal/processing/ory"
	"github.com/kyma-incubator/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// APIRuleReconciler reconciles a APIRule object
type APIRuleReconciler struct {
	client.Client
	Log                    logr.Logger
	OathkeeperSvc          string
	OathkeeperSvcPort      uint32
	CorsConfig             *processing.CorsConfig
	GeneratedObjectsLabels map[string]string
	ServiceBlockList       map[string][]string
	DomainAllowList        []string
	HostBlockList          []string
	DefaultDomainName      string
	Scheme                 *runtime.Scheme
	Config                 *helpers.Config
}

const (
	CONFIGMAP_NAME = "api-gateway-config"
	CONFIGMAP_NS   = "kyma-system"
)

type configMapPredicate struct {
	Log logr.Logger
	predicate.Funcs
}

func (p configMapPredicate) Create(e event.CreateEvent) bool {
	return p.Generic(event.GenericEvent(e))
}

func (p configMapPredicate) DeleteFunc(e event.DeleteEvent) bool {
	return p.Generic(event.GenericEvent{
		Object: e.Object,
	})
}

func (p configMapPredicate) Update(e event.UpdateEvent) bool {
	return p.Generic(event.GenericEvent{
		Object: e.ObjectNew,
	})
}

func (p configMapPredicate) Generic(e event.GenericEvent) bool {
	if e.Object == nil {
		p.Log.Error(nil, "Generic event has no object", "event", e)
		return false
	}
	_, okAP := e.Object.(*gatewayv1beta1.APIRule)
	configMap, okCM := e.Object.(*corev1.ConfigMap)
	return okAP || (okCM && configMap.GetNamespace() == CONFIGMAP_NS && configMap.GetName() == CONFIGMAP_NAME)
}

//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oathkeeper.ory.sh,resources=rules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=authorizationpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=security.istio.io,resources=requestauthentications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *APIRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Starting reconcilation")

	validator := validation.APIRule{
		ServiceBlockList:  r.ServiceBlockList,
		DomainAllowList:   r.DomainAllowList,
		HostBlockList:     r.HostBlockList,
		DefaultDomainName: r.DefaultDomainName,
	}
	r.Log.Info("Checking if it's ConfigMap reconciliation")
	isCMReconcile := (req.NamespacedName.String() == types.NamespacedName{Namespace: helpers.CM_NS, Name: helpers.CM_NAME}.String())
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
			r.Log.Info("ConfigMap changed")
			if len(configValidationFailures) > 0 {
				failuresJson, _ := json.Marshal(configValidationFailures)
				r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "Api", "failures": %s}`, string(failuresJson)))
			}
			return doneReconcile()
		}
	}
	r.Log.Info("Starting ApiRule reconciliation")
	apiRule := &gatewayv1beta1.APIRule{}

	err := r.Client.Get(ctx, req.NamespacedName, apiRule)
	if err != nil {
		r.Log.Error(err, "Error getting ApiRule")
		if apierrs.IsNotFound(err) {
			//There is no APIRule. Nothing to process, dependent objects will be garbage-collected.
			return doneReconcile()
		}

		//Nothing is yet processed: StatusSkipped
		status := processing.GetStatusForError(&r.Log, err, gatewayv1beta1.StatusSkipped)
		return r.updateStatusOrRetry(ctx, apiRule, status)
	}
	r.Log.Info("Validating ApiRule config")
	configValidationFailures := validator.ValidateConfig(r.Config)
	if len(configValidationFailures) > 0 {
		failuresJson, _ := json.Marshal(configValidationFailures)
		r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, apiRule.Namespace, apiRule.Name, string(failuresJson)))
		return r.updateStatusOrRetry(ctx, apiRule, processing.GetFailedValidationStatus(configValidationFailures))
	}

	//Prevent reconciliation after status update. It should be solved by controller-runtime implementation but still isn't.

	r.Log.Info("Validating if not status update or temporary annotation set")
	if apiRule.Generation != apiRule.Status.ObservedGeneration {
		r.Log.Info("not a status update")
		c := processing.ReconciliationConfig{
			OathkeeperSvc:     r.OathkeeperSvc,
			OathkeeperSvcPort: r.OathkeeperSvcPort,
			CorsConfig:        r.CorsConfig,
			AdditionalLabels:  r.GeneratedObjectsLabels,
			DefaultDomainName: r.DefaultDomainName,
			ServiceBlockList:  r.ServiceBlockList,
			DomainAllowList:   r.DomainAllowList,
			HostBlockList:     r.HostBlockList,
		}

		cmd := r.getReconciliation(c)
		r.Log.Info("Process reconcile")
		status := processing.Reconcile(ctx, r.Client, &r.Log, cmd, apiRule)
		r.Log.Info("Update status or retry")
		return r.updateStatusOrRetry(ctx, apiRule, status)
	}
	r.Log.Info("Finishing reconciliation")
	return doneReconcile()
}

func (r *APIRuleReconciler) getReconciliation(config processing.ReconciliationConfig) processing.ReconciliationCommand {
	if r.Config.JWTHandler == helpers.JWT_HANDLER_ISTIO {
		return istio.NewIstioReconciliation(config)
	}
	return ory.NewOryReconciliation(config)

}

// SetupWithManager sets up the controller with the Manager.
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1beta1.APIRule{}).
		Watches(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{}).
		WithEventFilter(&configMapPredicate{Log: r.Log}).
		Complete(r)
}

// Updates api status. If there was an error during update, returns the error so that entire reconcile loop is retried. If there is no error, returns a "reconcile success" value.
func (r *APIRuleReconciler) updateStatusOrRetry(ctx context.Context, api *gatewayv1beta1.APIRule, status processing.ReconciliationStatus) (ctrl.Result, error) {
	_, updateStatusErr := r.updateStatus(ctx, api, status)
	if updateStatusErr != nil {
		return retryReconcile(updateStatusErr) //controller retries to set the correct status eventually.
	}
	//Fail fast: If status is updated, users are informed about the problem. We don't need to reconcile again.
	return doneReconcile()
}

func doneReconcile() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func retryReconcile(err error) (ctrl.Result, error) {
	return reconcile.Result{Requeue: true}, err
}

func (r *APIRuleReconciler) updateStatus(ctx context.Context, api *gatewayv1beta1.APIRule, status processing.ReconciliationStatus) (*gatewayv1beta1.APIRule, error) {
	api.Status.ObservedGeneration = api.Generation
	api.Status.LastProcessedTime = &v1.Time{Time: time.Now()}
	api.Status.APIRuleStatus = status.ApiRuleStatus
	api.Status.VirtualServiceStatus = status.VirtualServiceStatus
	api.Status.AccessRuleStatus = status.AccessRuleStatus
	api.Status.RequestAuthenticationStatus = status.RequestAuthenticationStatus
	api.Status.AuthorizationPolicyStatus = status.AuthorizationPolicyStatus

	err := r.Client.Status().Update(ctx, api)
	if err != nil {
		return nil, err
	}
	return api, nil
}
