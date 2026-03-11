package gateway

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errReconcileWithBackoff = errors.New("reconciliation finished with error status")

func doneReconcileNoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func doneReconcileDefaultRequeue(reconcilerPeriod time.Duration) (ctrl.Result, error) {
	after := defaultReconciliationPeriod
	if reconcilerPeriod != 0 {
		after = reconcilerPeriod
	}

	return ctrl.Result{RequeueAfter: after}, nil
}

func doneReconcileErrorRequeue(err error, reconcilerPeriod time.Duration) (ctrl.Result, error) {
	after := errorReconciliationPeriod
	if reconcilerPeriod >= 0 {
		after = reconcilerPeriod
	}
	return ctrl.Result{RequeueAfter: after}, err
}

func doneReconcileMigrationRequeue(reconcilerPeriod time.Duration) (ctrl.Result, error) {
	after := migrationReconciliationPeriod
	if reconcilerPeriod != 0 {
		after = reconcilerPeriod
	}
	return ctrl.Result{RequeueAfter: after}, nil
}

func (r *APIRuleReconciler) updateStatus(ctx context.Context, l logr.Logger,
	apiRule client.Object, reconcileError bool) (ctrl.Result, error) {
	l.Info("Updating APIRule status")
	if err := r.Status().Update(ctx, apiRule); err != nil {
		l.Error(err, "Error updating APIRule status")
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}
	if _, ok := apiRule.GetAnnotations()[migration.AnnotationName]; ok {
		l.Info("Finished reconciliation", "next", r.MigrationReconcilePeriod)
		return doneReconcileMigrationRequeue(r.MigrationReconcilePeriod)
	}
	if reconcileError {
		l.Info("Finished reconciliation with error, will retry with backoff")
		return doneReconcileErrorRequeue(errReconcileWithBackoff, 0)
	}

	l.Info("Finished reconciliation", "next", r.ReconcilePeriod)
	return doneReconcileDefaultRequeue(r.ReconcilePeriod)
}
