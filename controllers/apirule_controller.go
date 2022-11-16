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
	"fmt"
	"time"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"

	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"

	"github.com/go-logr/logr"
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

		c := processing.NewReconciliationConfig(ctx, r)
		cmd := getReconciliation(c, "ory")

		status, validationStatus, err := processing.Reconcile(cmd, api)

		if validationStatus != nil {
			return r.setStatus(ctx, api, validationStatus, gatewayv1beta1.StatusSkipped)
		}

		if err != nil {
			return r.setStatusForError(ctx, api, err, status)
		}

		if status != gatewayv1beta1.StatusOK {
			return r.setStatusForError(ctx, api, err, status)
		}

		APIStatus := &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusOK,
		}

		return r.setStatus(ctx, api, APIStatus, gatewayv1beta1.StatusOK)
	}

	return doneReconcile()
}

func getReconciliation(config processing.ReconciliationConfig, featureFlag string) processing.ReconciliationCommand {
	if featureFlag == "istio" {
		return processing.NewIstioReconciliation(config)
	} else {
		return processing.NewOryReconciliation(config)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1beta1.APIRule{}).
		Complete(r)
}

// Sets status of APIRule. Accepts an auxilary status code that is used to report VirtualService and AccessRule status.
func (r *APIRuleReconciler) setStatus(ctx context.Context, api *gatewayv1beta1.APIRule, apiStatus *gatewayv1beta1.APIRuleResourceStatus, auxStatusCode gatewayv1beta1.StatusCode) (ctrl.Result, error) {
	virtualServiceStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	accessRuleStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	return r.updateStatusOrRetry(ctx, api, apiStatus, virtualServiceStatus, accessRuleStatus)
}

// Sets status of APIRule in error condition. Accepts an auxilary status code that is used to report VirtualService and AccessRule status.
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

// Updates api status. If there was an error during update, returns the error so that entire reconcile loop is retried. If there is no error, returns a "reconcile success" value.
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

func GenerateValidationStatus(failures []validation.Failure) *gatewayv1beta1.APIRuleResourceStatus {
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
