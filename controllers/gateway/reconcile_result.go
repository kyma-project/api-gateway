package gateway

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing/processors/migration"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxBackoff                  = time.Minute * 60
	maxRetryAttemptsBeforeReset = 10
)

func getObjectKey(obj client.Object) string {
	return fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName())
}

func (r *APIRuleReconciler) GetObjectRetryAttempt(obj client.Object) int {
	r.retryAttemptsMutex.RLock()
	defer r.retryAttemptsMutex.RUnlock()
	return r.retryAttempts[getObjectKey(obj)]
}

func (r *APIRuleReconciler) GetAllRetryAttempts() map[string]int {
	r.retryAttemptsMutex.RLock()
	defer r.retryAttemptsMutex.RUnlock()

	attempts := make(map[string]int, len(r.retryAttempts))
	for key, attempt := range r.retryAttempts {
		attempts[key] = attempt
	}
	return attempts
}

func (r *APIRuleReconciler) incrementRetryAttempt(obj client.Object) int {
	r.retryAttemptsMutex.Lock()
	defer r.retryAttemptsMutex.Unlock()

	key := getObjectKey(obj)
	newAttempt := r.retryAttempts[key] + 1
	if newAttempt > maxRetryAttemptsBeforeReset {
		newAttempt = maxRetryAttemptsBeforeReset
	}

	r.retryAttempts[key] = newAttempt
	return newAttempt
}

func (r *APIRuleReconciler) resetRetryAttempt(obj client.Object) {
	r.retryAttemptsMutex.Lock()
	defer r.retryAttemptsMutex.Unlock()
	delete(r.retryAttempts, getObjectKey(obj))
}

func calculateExponentialBackoff(basePeriod time.Duration, retryAttempt int) time.Duration {
	if retryAttempt < 1 {
		retryAttempt = 1
	}
	if retryAttempt > maxRetryAttemptsBeforeReset {
		retryAttempt = maxRetryAttemptsBeforeReset
	}
	calculated := time.Duration(math.Pow(2, float64(retryAttempt-1))) * basePeriod
	if calculated > maxBackoff {
		return maxBackoff
	}
	return calculated
}

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
	if reconcilerPeriod != 0 {
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
		newRetryAttempt := r.incrementRetryAttempt(apiRule)
		backoffPeriod := calculateExponentialBackoff(r.OnErrorReconcilePeriod, newRetryAttempt)
		l.Info("Finished reconciliation with error", "next", backoffPeriod)
		return doneReconcileErrorRequeue(nil, backoffPeriod)
	}

	r.resetRetryAttempt(apiRule)
	l.Info("Finished reconciliation", "next", r.ReconcilePeriod)
	return doneReconcileDefaultRequeue(r.ReconcilePeriod)
}
