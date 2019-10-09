/*

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

	"knative.dev/pkg/apis/istio/v1alpha3"

	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//APIReconciler reconciles a Api object
type APIReconciler struct {
	Client            client.Client
	Log               logr.Logger
	OathkeeperSvc     string
	OathkeeperSvcPort uint32
	JWKSURI           string
	Validator         APIRuleValidator
	CorsConfig        *processing.CorsConfig
}

//APIRuleValidator allows to validate APIRule instances created by the user.
type APIRuleValidator interface {
	Validate(apiRule *gatewayv1alpha1.APIRule, vsList v1alpha3.VirtualServiceList) []validation.Failure
}

//Reconcile .
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oathkeeper.ory.sh,resources=rules,verbs=get;list;watch;create;update;patch;delete
func (r *APIReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("Api", req.NamespacedName)

	api := &gatewayv1alpha1.APIRule{}

	err := r.Client.Get(ctx, req.NamespacedName, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			//There is no APIRule. Nothing to process, dependent objects will be garbage-collected.
			return doneReconcile()
		}

		//Nothing is yet processed: StatusSkipped
		return r.setStatusForError(ctx, api, err, gatewayv1alpha1.StatusSkipped)
	}

	//Prevent reconciliation after status update. It should be solved by controller-runtime implementation but still isn't.
	if api.Generation != api.Status.ObservedGeneration {

		//1.1) Get the list of existing Virtual Services to validate host
		var vsList v1alpha3.VirtualServiceList
		if err := r.Client.List(ctx, &vsList); err != nil {
			//Nothing is yet processed: StatusSkipped
			return r.setStatusForError(ctx, api, err, gatewayv1alpha1.StatusSkipped)
		}

		//1.2) Validate input including host
		validationFailures := r.Validator.Validate(api, vsList)
		if len(validationFailures) > 0 {
			r.Log.Info(fmt.Sprintf(`Validation failure {"controller": "Api", "request": "%s/%s"}`, api.Namespace, api.Name))
			return r.setStatus(ctx, api, generateValidationStatus(validationFailures), gatewayv1alpha1.StatusSkipped)
		}

		//2) Compute list of required objects (the set of objects required to satisfy our contract on apiRule.Spec, not yet applied)
		factory := processing.NewFactory(r.Client, r.Log, r.OathkeeperSvc, r.OathkeeperSvcPort, r.JWKSURI, r.CorsConfig)
		requiredObjects := factory.CalculateRequiredState(api)

		//3.1 Fetch all existing objects related to _this_ apiRule from the cluster (VS, Rules)
		actualObjects, err := factory.GetActualState(ctx, api)
		if err != nil {
			return r.setStatusForError(ctx, api, err, gatewayv1alpha1.StatusSkipped)
		}

		//3.2 Compute patch object
		patch := factory.CalculateDiff(requiredObjects, actualObjects)

		//3.3 Apply changes to the cluster
		err = factory.ApplyDiff(ctx, patch)
		if err != nil {
			//We don't know exactly which object(s) are not updated properly.
			//The safest approach is to assume nothing is correct and just use `StatusError`.
			return r.setStatusForError(ctx, api, err, gatewayv1alpha1.StatusError)
		}

		//4) Update status of CR
		APIStatus := &gatewayv1alpha1.APIRuleResourceStatus{
			Code: gatewayv1alpha1.StatusOK,
		}

		return r.setStatus(ctx, api, APIStatus, gatewayv1alpha1.StatusOK)
	}

	return doneReconcile()
}

//Sets status of APIRule. Accepts an auxilary status code that is used to report VirtualService and AccessRule status.
func (r *APIReconciler) setStatus(ctx context.Context, api *gatewayv1alpha1.APIRule, apiStatus *gatewayv1alpha1.APIRuleResourceStatus, auxStatusCode gatewayv1alpha1.StatusCode) (ctrl.Result, error) {
	virtualServiceStatus := &gatewayv1alpha1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	accessRuleStatus := &gatewayv1alpha1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	return r.updateStatusOrRetry(ctx, api, apiStatus, virtualServiceStatus, accessRuleStatus)
}

//Sets status of APIRule in error condition. Accepts an auxilary status code that is used to report VirtualService and AccessRule status.
func (r *APIReconciler) setStatusForError(ctx context.Context, api *gatewayv1alpha1.APIRule, err error, auxStatusCode gatewayv1alpha1.StatusCode) (ctrl.Result, error) {
	r.Log.Error(err, "Error during reconciliation")

	virtualServiceStatus := &gatewayv1alpha1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}
	accessRuleStatus := &gatewayv1alpha1.APIRuleResourceStatus{
		Code: auxStatusCode,
	}

	return r.updateStatusOrRetry(ctx, api, generateErrorStatus(err), virtualServiceStatus, accessRuleStatus)
}

//Updates api status. If there was an error during update, returns the error so that entire reconcile loop is retried. If there is no error, returns a "reconcile success" value.
func (r *APIReconciler) updateStatusOrRetry(ctx context.Context, api *gatewayv1alpha1.APIRule, apiStatus, virtualServiceStatus, accessRuleStatus *gatewayv1alpha1.APIRuleResourceStatus) (ctrl.Result, error) {
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

//SetupWithManager .
func (r *APIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha1.APIRule{}).
		Complete(r)
}

func (r *APIReconciler) updateStatus(ctx context.Context, api *gatewayv1alpha1.APIRule, APIStatus, virtualServiceStatus, accessRuleStatus *gatewayv1alpha1.APIRuleResourceStatus) (*gatewayv1alpha1.APIRule, error) {
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

func generateErrorStatus(err error) *gatewayv1alpha1.APIRuleResourceStatus {
	return toStatus(gatewayv1alpha1.StatusError, err.Error())
}

func generateValidationStatus(failures []validation.Failure) *gatewayv1alpha1.APIRuleResourceStatus {
	return toStatus(gatewayv1alpha1.StatusError, generateValidationDescription(failures))
}

func toStatus(c gatewayv1alpha1.StatusCode, desc string) *gatewayv1alpha1.APIRuleResourceStatus {
	return &gatewayv1alpha1.APIRuleResourceStatus{
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
