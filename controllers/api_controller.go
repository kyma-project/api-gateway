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
	"os"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

// ApiReconciler reconciles a Api object
type ApiReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apis/status,verbs=get;update;patch

func (r *ApiReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("api", req.NamespacedName)

	// demo sample fetching virtualservices

	list := networkingv1alpha3.VirtualServiceList{}
	err := r.Client.List(context.TODO(), &list, client.InNamespace(req.Namespace))
	if err != nil {
		fmt.Printf("ooops, error occured when fetching vs " + err.Error())
		os.Exit(1)
	}

	fmt.Println(list)

	return ctrl.Result{}, nil
}

func (r *ApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv2alpha1.Api{}).
		Complete(r)
}
