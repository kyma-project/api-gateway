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
	"github.com/kyma-incubator/api-gateway/internal/processing/ory"
	"time"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"

	"github.com/go-logr/logr"
	"github.com/kyma-incubator/api-gateway/internal/processing"
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

	apiRule := &gatewayv1beta1.APIRule{}

	err := r.Client.Get(ctx, req.NamespacedName, apiRule)
	if err != nil {
		if apierrs.IsNotFound(err) {
			//There is no APIRule. Nothing to process, dependent objects will be garbage-collected.
			return doneReconcile()
		}

		//Nothing is yet processed: StatusSkipped
		status := processing.GetStatusForError(&r.Log, err, gatewayv1beta1.StatusSkipped)
		return r.updateStatusOrRetry(ctx, apiRule, status)
	}

	//Prevent reconciliation after status update. It should be solved by controller-runtime implementation but still isn't.
	if apiRule.Generation != apiRule.Status.ObservedGeneration {

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

		cmd := getReconciliation(c)

		status := processing.Reconcile(ctx, r.Client, &r.Log, cmd, apiRule)

		return r.updateStatusOrRetry(ctx, apiRule, status)
	}

	return doneReconcile()
}

func getReconciliation(config processing.ReconciliationConfig) processing.ReconciliationCommand {
	// This should be replaced by the feature flag handling to return the appropriate reconciliation.
	return ory.NewOryReconciliation(config)
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1beta1.APIRule{}).
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

	err := r.Client.Status().Update(ctx, api)
	if err != nil {
		return nil, err
	}
	return api, nil
}
