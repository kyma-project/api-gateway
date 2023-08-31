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

package operator

import (
	"context"
	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewAPIGatewayReconciler(mgr manager.Manager) *APIGatewayReconciler {
	return &APIGatewayReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		log:           mgr.GetLogger(),
		statusHandler: newStatusHandler(mgr.GetClient()),
	}
}

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=apigateways/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the APIGateway object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *APIGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	apiGatewayCR := operatorv1alpha1.APIGateway{}
	if err := r.Client.Get(ctx, req.NamespacedName, &apiGatewayCR); err != nil {
		if apierrors.IsNotFound(err) {
			r.log.Info("Skipped reconciliation, because ApiGateway CR was not found")
			return ctrl.Result{}, nil
		}
		r.log.Info("Could not get istio CR")
		return ctrl.Result{}, err
	}

	if err := r.statusHandler.updateToReady(ctx, &apiGatewayCR); err != nil {
		r.log.Error(err, "Update status to ready failed")
		return ctrl.Result{}, err
	} else {
		r.log.Info("Reconciled status successfully")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.APIGateway{}).
		Complete(r)
}
