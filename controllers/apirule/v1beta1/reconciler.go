// Package reconciler contains the APIRule v1beta1 reconciler.
package reconciler

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"kyma-project.io/api-gateway/internal/controllers/apirule/v1beta1/utils"
	"kyma-project.io/api-gateway/internal/logging"
	"kyma-project.io/api-gateway/internal/metrics"
	"kyma-project.io/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	reconciliationDisabled = true
)

// Reconciler is the APIRule v1beta1 reconciler.
type Reconciler struct {
	// ... existing fields ...
	reconciliationDisabled bool
}

// NewReconciler creates a new APIRule v1beta1 reconciler.
func NewReconciler(client client.Client, config *config.Config, recorder record.EventRecorder, metricsRecorder metrics.Recorder) (*Reconciler, error) {
	// ... existing code ...
	reconciler.reconciliationDisabled = reconciliationDisabled
	return reconciler, nil
}

// Reconcile is the reconcile function for the APIRule v1beta1 reconciler.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r.reconciliationDisabled {
		return reconcile.Result{}, nil
	}
	// ... existing code ...
}
