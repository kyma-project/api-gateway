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

	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// APIRuleReconciler reconciles a APIRule object
type APIRuleReconciler struct {
	client.Client
	Log                    logr.Logger
	OathkeeperSvc          string
	OathkeeperSvcPort      uint32
	JWKSURI                string
	CorsConfig             *processing.CorsConfig
	GeneratedObjectsLabels map[string]string
	ServiceBlockList       map[string][]string
	DomainAllowList        []string
	HostBlockList          []string
	DefaultDomainName      string
	Scheme                 *runtime.Scheme
}

//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oathkeeper.ory.sh,resources=rules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apiextensions.k8s.io",resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete

func (r *APIRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = r.Log.WithValues("Api", req.NamespacedName)

	r.Log.Info("Starting reconcilation")

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

	//Prevent reconciliation after status update. It should be solved by controller-runtime implementation but still isn't.
	if api.Generation != api.Status.ObservedGeneration {

		//1.1) Get the list of existing Virtual Services to validate host
		var vsList networkingv1beta1.VirtualServiceList
		if err := r.Client.List(ctx, &vsList); err != nil {
			//Nothing is yet processed: StatusSkipped
			return r.setStatusForError(ctx, api, err, gatewayv1beta1.StatusSkipped)
		}

		//1.2) Validate input including host
		validator := validation.APIRule{
			ServiceBlockList:  r.ServiceBlockList,
			DomainAllowList:   r.DomainAllowList,
			HostBlockList:     r.HostBlockList,
			DefaultDomainName: r.DefaultDomainName,
		}
		validationFailures := validator.Validate(api, vsList)
		if len(validationFailures) > 0 {
			failuresJson, _ := json.Marshal(validationFailures)
			r.Log.Info(fmt.Sprintf(`Validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, api.Namespace, api.Name, string(failuresJson)))
			return r.setStatus(ctx, api, generateValidationStatus(validationFailures), gatewayv1beta1.StatusSkipped)
		}

		//2) Compute list of required objects (the set of objects required to satisfy our contract on apiRule.Spec, not yet applied)
		factory := processing.NewFactory(r.Client, r.Log, r.OathkeeperSvc, r.OathkeeperSvcPort, r.JWKSURI, r.CorsConfig, r.GeneratedObjectsLabels, r.DefaultDomainName)
		requiredObjects := factory.CalculateRequiredState(api)

		//3.1 Fetch all existing objects related to _this_ apiRule from the cluster (VS, Rules)
		actualObjects, err := factory.GetActualState(ctx, api)
		if err != nil {
			return r.setStatusForError(ctx, api, err, gatewayv1beta1.StatusSkipped)
		}

		//3.2 Compute patch object
		patch := factory.CalculateDiff(requiredObjects, actualObjects)

		//3.3 Apply changes to the cluster
		err = factory.ApplyDiff(ctx, patch)
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
		Complete(r)
}

//Sets status of APIRule. Accepts an auxilary status code that is used to report VirtualService and AccessRule status.
func (r *APIRuleReconciler) setStatus(ctx context.Context, api *gatewayv1beta1.APIRule, apiStatus *gatewayv1beta1.APIRuleResourceStatus, auxStatusCode gatewayv1beta1.StatusCode) (ctrl.Result, error) {
	virtualServiceStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	accessRuleStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	return r.updateStatusOrRetry(ctx, api, apiStatus, virtualServiceStatus, accessRuleStatus)
}

//Sets status of APIRule in error condition. Accepts an auxilary status code that is used to report VirtualService and AccessRule status.
func (r *APIRuleReconciler) setStatusForError(ctx context.Context, api *gatewayv1beta1.APIRule, err error, auxStatusCode gatewayv1beta1.StatusCode) (ctrl.Result, error) {
	r.Log.Error(err, "Error during reconciliation")

	virtualServiceStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	accessRuleStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}

	return r.updateStatusOrRetry(ctx, api, generateErrorStatus(err), virtualServiceStatus, accessRuleStatus)
}

//Updates api status. If there was an error during update, returns the error so that entire reconcile loop is retried. If there is no error, returns a "reconcile success" value.
func (r *APIRuleReconciler) updateStatusOrRetry(ctx context.Context, api *gatewayv1beta1.APIRule, apiStatus, virtualServiceStatus, accessRuleStatus *gatewayv1beta1.APIRuleResourceStatus) (ctrl.Result, error) {
	_, updateStatusErr := r.updateStatus(ctx, api, apiStatus, virtualServiceStatus, accessRuleStatus)
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

func (r *APIRuleReconciler) updateStatus(ctx context.Context, api *gatewayv1beta1.APIRule, APIStatus, virtualServiceStatus, accessRuleStatus *gatewayv1beta1.APIRuleResourceStatus) (*gatewayv1beta1.APIRule, error) {
	api.Status.ObservedGeneration = api.Generation
	api.Status.LastProcessedTime = &v1.Time{Time: time.Now()}
	api.Status.APIRuleStatus = APIStatus
	api.Status.VirtualServiceStatus = virtualServiceStatus
	api.Status.AccessRuleStatus = accessRuleStatus

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
