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

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"

	"github.com/kyma-incubator/api-gateway/internal/helpers"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

	isCMReconcile := (req.NamespacedName.String() == types.NamespacedName{Namespace: helpers.CM_NS, Name: helpers.CM_NAME}.String())
	if isCMReconcile || r.Config.JWTHandler == "" {
		err := r.Config.ReadFromConfigMap(ctx, r.Client)
		if err != nil {
			if apierrs.IsNotFound(err) {
				r.Log.Info( fmt.Sprintf(`ConfigMap %s in namespace %s was not found {"controller": "Api"}, will use default config`, helpers.CM_NAME, helpers.CM_NS))
				r.Config.ResetToDefault()
			} else {
				r.Log.Error(err, fmt.Sprintf(`could not read ConfigMap %s in namespace %s {"controller": "Api"}`, helpers.CM_NAME, helpers.CM_NS))
				r.Config.Reset()
			}
		}
		if isCMReconcile {
			return doneReconcile()
		}
	}

	api := &gatewayv1beta1.APIRule{}
	err := r.Client.Get(ctx, req.NamespacedName, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			//There is no APIRule. Nothing to process, dependent objects will be garbage-collected.
			return doneReconcile()
		}

		//Nothing is yet processed: StatusSkipped
		return r.setStatusForError(ctx, api, err, gatewayv1beta1.StatusSkipped)
	}

	validator := validation.APIRule{
		ServiceBlockList:  r.ServiceBlockList,
		DomainAllowList:   r.DomainAllowList,
		HostBlockList:     r.HostBlockList,
		DefaultDomainName: r.DefaultDomainName,
	}

	configValidationFailures := validator.ValidateConfig(r.Config)
	if len(configValidationFailures) > 0 {
		failuresJson, _ := json.Marshal(configValidationFailures)
		r.Log.Error(err, fmt.Sprintf(`Config validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, api.Namespace, api.Name, string(failuresJson)))
		return r.setStatus(ctx, api, generateValidationStatus(configValidationFailures), gatewayv1beta1.StatusError)
	}

	//Prevent reconciliation after status update. It should be solved by controller-runtime implementation but still isn't.
	if api.Generation != api.Status.ObservedGeneration {
		//1.1) Get the list of existing Virtual Services to validate host
		var vsList networkingv1beta1.VirtualServiceList
		if err := r.Client.List(ctx, &vsList); err != nil {
			//Nothing is yet processed: StatusSkipped
			return r.setStatusForError(ctx, api, err, gatewayv1beta1.StatusSkipped)
		}

		//1.2) Validate input including host
		validationFailures := validator.Validate(api, vsList, r.Config)
		if len(validationFailures) > 0 {
			failuresJson, _ := json.Marshal(validationFailures)
			r.Log.Info(fmt.Sprintf(`Validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, api.Namespace, api.Name, string(failuresJson)))
			return r.setStatus(ctx, api, generateValidationStatus(validationFailures), gatewayv1beta1.StatusSkipped)
		}

		//2) Compute list of required objects (the set of objects required to satisfy our contract on apiRule.Spec, not yet applied)
		factory := processing.NewFactory(r.Client, r.Log, r.OathkeeperSvc, r.OathkeeperSvcPort, r.CorsConfig, r.GeneratedObjectsLabels, r.DefaultDomainName)
		requiredObjects := factory.CalculateRequiredState(api, r.Config)

		//3.1 Fetch all existing objects related to _this_ apiRule from the cluster (VS, Rules)
		actualObjects, err := factory.GetActualState(ctx, api, r.Config)
		if err != nil {
			return r.setStatusForError(ctx, api, err, gatewayv1beta1.StatusSkipped)
		}

		//3.2 Compute patch object
		patch := factory.CalculateDiff(requiredObjects, actualObjects, r.Config)

		//3.3 Apply changes to the cluster
		err = factory.ApplyDiff(ctx, patch, r.Config)
		if err != nil {
			//We don't know exactly which object(s) are not updated properly.
			//The safest approach is to assume nothing is correct and just use `StatusError`.
			return r.setStatusForError(ctx, api, err, gatewayv1beta1.StatusError)
		}

		//4) Update status of CR
		APIStatus := &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusOK,
		}

		return r.setStatus(ctx, api, APIStatus, gatewayv1beta1.StatusOK)
	}

	return doneReconcile()
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1beta1.APIRule{}).
		Owns(&corev1.ConfigMap{}).
		WithEventFilter(&configMapPredicate{Log: r.Log}).
		Complete(r)
}

// Sets status of APIRule. Accepts an auxilary status code that is used to report VirtualService, AccessRule, RequestAuthentication and AuthorizationPolicy statuses.
func (r *APIRuleReconciler) setStatus(ctx context.Context, api *gatewayv1beta1.APIRule, apiStatus *gatewayv1beta1.APIRuleResourceStatus, auxStatusCode gatewayv1beta1.StatusCode) (ctrl.Result, error) {
	virtualServiceStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	accessRuleStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	reqAuthStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	authPolicyStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}

	return r.updateStatusOrRetry(ctx, api, apiStatus, virtualServiceStatus, accessRuleStatus, reqAuthStatus, authPolicyStatus)
}

// Sets status of APIRule in error condition. Accepts an auxilary status code that is used to report VirtualService, AccessRule, RequestAuthentication and AuthorizationPolicy statuses.
func (r *APIRuleReconciler) setStatusForError(ctx context.Context, api *gatewayv1beta1.APIRule, err error, auxStatusCode gatewayv1beta1.StatusCode) (ctrl.Result, error) {
	r.Log.Error(err, "Error during reconciliation")

	virtualServiceStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	accessRuleStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	reqAuthStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	authPolicyStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}

	return r.updateStatusOrRetry(ctx, api, generateErrorStatus(err), virtualServiceStatus, accessRuleStatus, reqAuthStatus, authPolicyStatus)
}

// Updates api status. If there was an error during update, returns the error so that entire reconcile loop is retried. If there is no error, returns a "reconcile success" value.
func (r *APIRuleReconciler) updateStatusOrRetry(ctx context.Context, api *gatewayv1beta1.APIRule, apiStatus, virtualServiceStatus, accessRuleStatus, reqAuthStatus, authPolicyStatus *gatewayv1beta1.APIRuleResourceStatus) (ctrl.Result, error) {
	_, updateStatusErr := r.updateStatus(ctx, api, apiStatus, virtualServiceStatus, accessRuleStatus, reqAuthStatus, authPolicyStatus)
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

func (r *APIRuleReconciler) updateStatus(ctx context.Context, api *gatewayv1beta1.APIRule, APIStatus, virtualServiceStatus, accessRuleStatus, reqAuthStatus, authPolicyStatus *gatewayv1beta1.APIRuleResourceStatus) (*gatewayv1beta1.APIRule, error) {
	api.Status.ObservedGeneration = api.Generation
	api.Status.LastProcessedTime = &v1.Time{Time: time.Now()}
	api.Status.APIRuleStatus = APIStatus
	api.Status.VirtualServiceStatus = virtualServiceStatus
	api.Status.AccessRuleStatus = accessRuleStatus
	api.Status.RequestAuthenticationStatus = reqAuthStatus
	api.Status.AuthorizationPolicyStatus = authPolicyStatus

	err := r.Client.Status().Update(ctx, api)
	if err != nil {
		return nil, err
	}
	return api, nil
}

func generateErrorStatus(err error) *gatewayv1beta1.APIRuleResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, err.Error())
}

func generateValidationStatus(failures []validation.Failure) *gatewayv1beta1.APIRuleResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, generateValidationDescription(failures))
}

func toStatus(c gatewayv1beta1.StatusCode, desc string) *gatewayv1beta1.APIRuleResourceStatus {
	return &gatewayv1beta1.APIRuleResourceStatus{
		Code:        c,
		Description: desc,
	}
}

func generateValidationDescription(failures []validation.Failure) string {
	var description string

	if len(failures) == 1 {
		description = "Validation error: "
		description += fmt.Sprintf("Attribute \"%s\": %s", failures[0].AttributePath, failures[0].Message)
	} else {
		const maxEntries = 3
		description = "Multiple validation errors: "
		for i := 0; i < len(failures) && i < maxEntries; i++ {
			description += fmt.Sprintf("\nAttribute \"%s\": %s", failures[i].AttributePath, failures[i].Message)
		}
		if len(failures) > maxEntries {
			description += fmt.Sprintf("\n%d more error(s)...", len(failures)-maxEntries)
		}
	}

	return description
}
