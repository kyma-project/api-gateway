package gateway

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

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

func doneReconcileErrorRequeue(reconcilerPeriod time.Duration) (ctrl.Result, error) {
	after := errorReconciliationPeriod
	if reconcilerPeriod != 0 {
		after = reconcilerPeriod
	}
	return ctrl.Result{RequeueAfter: after}, nil
}

func doneReconcileMigrationRequeue(reconcilerPeriod time.Duration) (ctrl.Result, error) {
	after := migrationReconciliationPeriod
	if reconcilerPeriod != 0 {
		after = reconcilerPeriod
	}
	return ctrl.Result{RequeueAfter: after}, nil
}

func (r *APIRuleReconciler) updateStatus(ctx context.Context, l logr.Logger, apiRule *gatewayv1beta1.APIRule, s status.ReconciliationStatus) (ctrl.Result, error) {
	// TODO handle v2alpha1
	err := s.UpdateStatus(status.Status{V1beta1Status: &apiRule.Status})
	if err != nil {
		l.Error(err, "Error updating APIRule status")
		// try again
		return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
	}

	// finally update object status
	l.Info("Updating APIRule status")
	if err := r.Status().Update(ctx, apiRule); err != nil {
		l.Error(err, "Error updating APIRule status")
		return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
	}
	if metav1.HasAnnotation(apiRule.ObjectMeta, migration.AnnotationName) {
		l.Info("Finished reconciliation", "next", r.MigrationReconcilePeriod)
		return doneReconcileMigrationRequeue(r.MigrationReconcilePeriod)
	}
	l.Info("Finished reconciliation", "next", r.ReconcilePeriod)
	return doneReconcileDefaultRequeue(r.ReconcilePeriod)
}
