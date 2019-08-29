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
	"time"

	"github.com/kyma-incubator/api-gateway/internal/processing"

	"github.com/go-logr/logr"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/clients"
	"github.com/kyma-incubator/api-gateway/internal/validation"
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
}

//Reconcile .
// +kubebuilder:rbac:groups=authentication.istio.io,resources=policies,verbs=get;list;watch;create;update;patch;delete
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
			return reconcile.Result{}, err
		}
	}

	APIStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code: gatewayv2alpha1.StatusOK,
	}

	virtualServiceStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.StatusSkipped,
		Description: "Skipped setting Istio Virtual Service",
	}
	policyStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.StatusSkipped,
		Description: "Skipped setting Istio Policy",
	}

	accessRuleStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.StatusSkipped,
		Description: "Skipped setting Oathkeeper Access Rule",
	}

	if api.Generation != api.Status.ObservedGeneration {
		r.Log.Info("Api processing")

		validationStrategy, err := validation.NewFactory(r.Log).StrategyFor(*api.Spec.Auth.Name)
		if err != nil {
			_, updateStatErr := r.updateStatus(ctx, api, generateErrorStatus(err), virtualServiceStatus, policyStatus, accessRuleStatus)
			if updateStatErr != nil {
				return reconcile.Result{Requeue: true}, err
			}
			return ctrl.Result{}, err
		}

		err = validationStrategy.Validate(api.Spec.Auth.Config)
		if err != nil {
			_, updateStatErr := r.updateStatus(ctx, api, generateErrorStatus(err), virtualServiceStatus, policyStatus, accessRuleStatus)
			if updateStatErr != nil {
				return reconcile.Result{Requeue: true}, err
			}
			return ctrl.Result{}, err
		}

		processingStrategy, err := processing.NewFactory(r.ExtCRClients.ForVirtualService(), r.ExtCRClients.ForAuthenticationPolicy(), r.ExtCRClients.ForAccessRule(), r.Log, r.OathkeeperSvc, r.OathkeeperSvcPort, r.JWKSURI).StrategyFor(*api.Spec.Auth.Name)
		if err != nil {
			_, updateStatErr := r.updateStatus(ctx, api, generateErrorStatus(err), virtualServiceStatus, policyStatus, accessRuleStatus)
			if updateStatErr != nil {
				return reconcile.Result{Requeue: true}, err
			}
			return ctrl.Result{}, err
		}

		err = processingStrategy.Process(ctx, api)
		if err != nil {
			virtualServiceStatus := &gatewayv2alpha1.GatewayResourceStatus{
				Code:        gatewayv2alpha1.StatusError,
				Description: err.Error(),
			}

			_, updateStatErr := r.updateStatus(ctx, api, generateErrorStatus(err), virtualServiceStatus, policyStatus, accessRuleStatus)
			if updateStatErr != nil {
				return reconcile.Result{Requeue: true}, err
			}
			return ctrl.Result{}, err
		}

		virtualServiceStatus := &gatewayv2alpha1.GatewayResourceStatus{
			Code: gatewayv2alpha1.StatusOK,
		}

		_, err = r.updateStatus(ctx, api, APIStatus, virtualServiceStatus, policyStatus, accessRuleStatus)

		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

//SetupWithManager .
func (r *APIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv2alpha1.Gate{}).
		Complete(r)
}

func (r *APIReconciler) updateStatus(ctx context.Context, api *gatewayv2alpha1.Gate, APIStatus, virtualServiceStatus, policyStatus, accessRuleStatus *gatewayv2alpha1.GatewayResourceStatus) (*gatewayv2alpha1.Gate, error) {
	api.Status.ObservedGeneration = api.Generation
	api.Status.LastProcessedTime = &v1.Time{Time: time.Now()}
	api.Status.GateStatus = APIStatus
	api.Status.VirtualServiceStatus = virtualServiceStatus
	api.Status.PolicyServiceStatus = policyStatus
	api.Status.AccessRuleStatus = accessRuleStatus

	err := r.Status().Update(ctx, api)
	if err != nil {
		return nil, err
	}
	return api, nil
}

func generateErrorStatus(err error) *gatewayv2alpha1.GatewayResourceStatus {
	return &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.StatusError,
		Description: err.Error(),
	}
}
