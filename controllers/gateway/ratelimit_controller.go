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
	ratelimitv1alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RateLimitReconciler reconciles a RateLimit object
type RateLimitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// There should be no kubebuilder:rbac markers in this file as it's hard to modify the ClusterRole rules array in
// kustomize. The roles are managed in the file config/dev/kustomization.yaml. Once this feature is ready for release,
// the markers can be added again.

func (r *RateLimitReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("###################### RateLimitReconciler.Reconcile ######################")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RateLimitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ratelimitv1alpha1.RateLimit{}).
		Complete(r)
}
