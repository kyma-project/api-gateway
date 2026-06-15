// Package v1beta1 contains the APIRule v1beta1 controller.
package v1beta1

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"kyma-project.io/api-gateway/internal/controllers/apirule/v1beta1/reconciler"
	"kyma-project.io/api-gateway/internal/controllers/apirule/v1beta1/utils"
	"kyma-project.io/api-gateway/internal/logging"
	"kyma-project.io/api-gateway/internal/metrics"
	"kyma-project.io/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	controllerName = "apirule-controller"
)

// Add creates a new APIRule v1beta1 controller.
func Add(mgr manager.Manager, recorder record.EventRecorder, config *config.Config) error {
	// Create a new reconciler.
	reconciler, err := reconciler.NewReconciler(mgr.GetClient(), config, recorder, metrics.NewRecorder())
	if err != nil {
		return err
	}
	// Create a new controller.
	ctrl, err := controller.NewUnmanaged(controllerName, mgr, reconciler)
	if err != nil {
		return err
	}
	// Add the controller to the manager.
	if err := mgr.Add(ctrl); err != nil {
		return err
	}
	// Disable reconciliation for APIRule v1beta1.
	ctrl.GetReconciler().DisableReconciliation()
	return nil
}
