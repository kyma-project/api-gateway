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
	"errors"
	"fmt"
	"time"

	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"

	"github.com/go-logr/logr"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/clients"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//APIReconciler reconciles a Api object
type APIReconciler struct {
	ExtCRClients *clients.ExternalCRClients
	client.Client
	Log               logr.Logger
	OathkeeperSvc     string
	OathkeeperSvcPort uint32
	JWKSURI           string
	Validator         GateValidator
}

//GateValidator allows to validate Gate instances created by the user.
type GateValidator interface {
	Validate(gate *gatewayv2alpha1.Gate) []validation.Failure
}

//Reconcile .
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=gates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=gates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oathkeeper.ory.sh,resources=rules,verbs=get;list;watch;create;update;patch;delete
func (r *APIReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("api", req.NamespacedName)

	api := &gatewayv2alpha1.Gate{}

	err := r.Get(ctx, req.NamespacedName, api)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return retryReconcile(err)
		}
	}

	APIStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code: gatewayv2alpha1.StatusOK,
	}

	virtualServiceStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.StatusSkipped,
		Description: "Skipped setting Istio Virtual Service",
	}

	accessRuleStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.StatusSkipped,
		Description: "Skipped setting Oathkeeper Access Rule",
	}

	//Prevent reconciliation after status update. It should be solved by controller-runtime implementation but still isn't.
	if api.Generation != api.Status.ObservedGeneration {

		validationFailures := r.Validator.Validate(api)
		if len(validationFailures) > 0 {
			r.Log.Info(fmt.Sprintf(`Validation failure {"controller": "gate", "request": "%s/%s"}`, api.Namespace, api.Name))
			gateValidationStatus := generateValidationStatus(validationFailures)
			_, updateStatErr := r.updateStatus(ctx, api, gateValidationStatus, virtualServiceStatus, accessRuleStatus)
			if updateStatErr != nil {
				r.Log.Error(errors.New("Status couldn't be updated with validation failures"), gateValidationStatus.Description)
				return retryReconcile(updateStatErr)
			}
			return doneReconcile()
		}

		err = processing.NewFactory(r.ExtCRClients.ForVirtualService(), r.ExtCRClients.ForAccessRule(), r.Log, r.OathkeeperSvc, r.OathkeeperSvcPort, r.JWKSURI).Run(ctx, api)
		if err != nil {
			virtualServiceStatus = &gatewayv2alpha1.GatewayResourceStatus{
				Code:        gatewayv2alpha1.StatusError,
				Description: err.Error(),
			}
			accessRuleStatus = &gatewayv2alpha1.GatewayResourceStatus{
				Code:        gatewayv2alpha1.StatusError,
				Description: err.Error(),
			}

			_, updateStatErr := r.updateStatus(ctx, api, generateErrorStatus(err), virtualServiceStatus, accessRuleStatus)
			if updateStatErr != nil {
				//Log original error (it will be silently lost otherwise)
				r.Log.Error(err, "Error during reconcilation")
				return retryReconcile(updateStatErr)
			}
			//Fail fast: If status is updated we don't want to reconcile again.
			return doneReconcile()
		}

		virtualServiceStatus = &gatewayv2alpha1.GatewayResourceStatus{
			Code: gatewayv2alpha1.StatusOK,
		}

		accessRuleStatus = &gatewayv2alpha1.GatewayResourceStatus{
			Code: gatewayv2alpha1.StatusOK,
		}

		_, err = r.updateStatus(ctx, api, APIStatus, virtualServiceStatus, accessRuleStatus)

		if err != nil {
			return retryReconcile(err)
		}
	}

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
		For(&gatewayv2alpha1.Gate{}).
		Complete(r)
}

func (r *APIReconciler) updateStatus(ctx context.Context, api *gatewayv2alpha1.Gate, APIStatus, virtualServiceStatus, accessRuleStatus *gatewayv2alpha1.GatewayResourceStatus) (*gatewayv2alpha1.Gate, error) {
	api.Status.ObservedGeneration = api.Generation
	api.Status.LastProcessedTime = &v1.Time{Time: time.Now()}
	api.Status.GateStatus = APIStatus
	api.Status.VirtualServiceStatus = virtualServiceStatus
	api.Status.AccessRuleStatus = accessRuleStatus

	err := r.Status().Update(ctx, api)
	if err != nil {
		return nil, err
	}
	return api, nil
}

func generateErrorStatus(err error) *gatewayv2alpha1.GatewayResourceStatus {
	return toStatus(gatewayv2alpha1.StatusError, err.Error())
}

func generateValidationStatus(failures []validation.Failure) *gatewayv2alpha1.GatewayResourceStatus {
	return toStatus(gatewayv2alpha1.StatusError, generateValidationDescription(failures))
}

func toStatus(c gatewayv2alpha1.StatusCode, desc string) *gatewayv2alpha1.GatewayResourceStatus {
	return &gatewayv2alpha1.GatewayResourceStatus{
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
