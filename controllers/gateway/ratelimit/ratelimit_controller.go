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

package ratelimit

import (
	"context"
	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/ratelimit/v1alpha1"
	"github.com/kyma-project/api-gateway/internal/ratelimit"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

const defaultReconciliationPeriod = 30 * time.Minute

// RateLimitReconciler reconciles a RateLimit object
type RateLimitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// There should be no kubebuilder:rbac markers in this file as it's hard to modify the ClusterRole rules array in
// kustomize. The roles are managed in the file config/dev/kustomization.yaml. Once this feature is ready for release,
// the markers can be added again.

// Reconcile is part of the main Kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// In this function, the RateLimit object is fetched and validated.
// If the object is not found, it is ignored. If validation fails, an error is returned.
// Otherwise, the function returns a result with a requeue period.
func (r *RateLimitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("namespace", req.Namespace, "RateLimit", req.Name)
	l.Info("Starting reconciliation")

	rateLimit := ratelimitv1alpha1.RateLimit{}
	if err := r.Get(ctx, req.NamespacedName, &rateLimit); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := ratelimit.Validate(ctx, r.Client, rateLimit)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: defaultReconciliationPeriod}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RateLimitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ratelimitv1alpha1.RateLimit{}).
		Complete(r)
}
