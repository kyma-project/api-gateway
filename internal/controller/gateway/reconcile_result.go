package gateway

import (
	"context"
	"errors"
	"time"

	"github.com/go-logr/logr"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
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

// updateStatus writes the APIRule status under a conflict-retry loop. Each attempt
// re-Gets the CR, re-applies this pass's desired State/Description/Conditions onto the
// fresh object, and (when removeStale is true) prunes condition types present on the CR
// but absent from desiredConditions. The terminal success write passes every condition
// the current spec produces, so a type missing there is stale; error and processing writes
// pass a partial set and must not prune. The write is skipped when the fresh object
// already matches, to avoid a no-op API call every reconcile.
//
// desiredConditions must be the pure set this pass produced — it is the prune authority.
// Deriving it from apiRule.Status.Conditions would be wrong: that field is the upsert-merge
// of the fresh set onto whatever stale conditions the cached object carried, so a no-longer-
// applicable type would still appear there and escape pruning.
func (r *APIRuleReconciler) updateStatus(ctx context.Context, l logr.Logger,
	apiRule client.Object, desiredConditions []metav1.Condition, reconcileError bool, removeStale bool) (ctrl.Result, error) {
	l.Info("Updating APIRule status")

	v2, ok := apiRule.(*gatewayv2alpha1.APIRule)
	if !ok {
		// No status-bearing non-v2 object reaches this path; fall back to a plain write.
		if err := r.Status().Update(ctx, apiRule); err != nil {
			l.Error(err, "Error updating APIRule status")
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}
		return r.finishStatusUpdate(l, apiRule, reconcileError)
	}

	// Snapshot the desired status before the loop; the re-Get inside each attempt
	// overwrites the CR's status with server state, so these must survive it.
	wantState := v2.Status.State
	wantDescription := v2.Status.Description
	wantConditions := append([]metav1.Condition(nil), desiredConditions...)

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := r.Get(ctx, client.ObjectKeyFromObject(v2), v2); getErr != nil {
			return getErr
		}

		prevState := v2.Status.State
		prevDescription := v2.Status.Description
		v2.Status.State = wantState
		v2.Status.Description = wantDescription

		changed := prevState != wantState || prevDescription != wantDescription
		for _, c := range wantConditions {
			if apimeta.SetStatusCondition(&v2.Status.Conditions, c) {
				changed = true
			}
		}
		if removeStale {
			written := make(map[string]struct{}, len(wantConditions))
			for _, c := range wantConditions {
				written[c.Type] = struct{}{}
			}
			for _, c := range append([]metav1.Condition(nil), v2.Status.Conditions...) {
				if _, ok := written[c.Type]; !ok {
					apimeta.RemoveStatusCondition(&v2.Status.Conditions, c.Type)
					changed = true
				}
			}
		}

		if !changed {
			return nil
		}
		return r.Status().Update(ctx, v2)
	})
	if err != nil {
		l.Error(err, "Error updating APIRule status")
		return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
	}

	return r.finishStatusUpdate(l, v2, reconcileError)
}

// finishStatusUpdate picks the requeue result after a successful status write.
func (r *APIRuleReconciler) finishStatusUpdate(l logr.Logger, apiRule client.Object, reconcileError bool) (ctrl.Result, error) {
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
