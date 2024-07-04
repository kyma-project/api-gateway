package gateway

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// Updates api status. If there was an error during update, returns the error so that entire reconcile loop is retried. If there is no error, returns a "reconcile success" value.
func (r *APIRuleReconciler) updateStatusOrRetry(ctx context.Context, api *gatewayv1beta1.APIRule, status status.ReconciliationStatus) (ctrl.Result, error) {
	_, updateStatusErr := r.updateStatus(ctx, api, status)
	if updateStatusErr != nil {
		r.Log.Error(updateStatusErr, "Error updating ApiRule status, retrying")
		return retryReconcile(updateStatusErr) //controller retries to set the correct status eventually.
	}

	// If error happened during reconciliation (e.g. VirtualService conflict) requeue for reconciliation earlier
	if status.HasError() {
		r.Log.Info("Requeue for reconciliation because the status has an error")
		return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
	}

	return doneReconcileDefaultRequeue(r.ReconcilePeriod, &r.Log)
}

func (r *APIRuleReconciler) updateStatus(ctx context.Context, api *gatewayv1beta1.APIRule, status status.ReconciliationStatus) (*gatewayv1beta1.APIRule, error) {
	api, err := r.getLatestApiRule(ctx, api)
	if err != nil {
		return nil, err
	}

	api.Status.ObservedGeneration = api.Generation
	api.Status.LastProcessedTime = &v1.Time{Time: time.Now()}

	err = status.UpdateStatus(&api.Status)
	if err != nil {
		return nil, err
	}

	r.Log.Info("Updating ApiRule status", "status", api.Status)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err = r.Client.Status().Update(ctx, api)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return api, nil
}

func (r *APIRuleReconciler) getLatestApiRule(ctx context.Context, api *gatewayv1beta1.APIRule) (*gatewayv1beta1.APIRule, error) {
	apiRule := &gatewayv1beta1.APIRule{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: api.Name, Namespace: api.Namespace}, apiRule)
	if err != nil {
		if apierrs.IsNotFound(err) {
			r.Log.Error(err, "ApiRule not found")
			return nil, err
		}

		r.Log.Error(err, "Error getting ApiRule")
		return nil, err
	}

	return apiRule, nil
}

func doneReconcileNoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func doneReconcileDefaultRequeue(reconcilerPeriod time.Duration, logger *logr.Logger) (ctrl.Result, error) {
	after := defaultReconciliationPeriod
	if reconcilerPeriod != 0 {
		after = reconcilerPeriod
	}

	logger.Info("Finished reconciliation and requeue", "requeue period", after)
	return ctrl.Result{RequeueAfter: after}, nil
}

func doneReconcileErrorRequeue(errorReconcilerPeriod time.Duration) (ctrl.Result, error) {
	after := errorReconciliationPeriod
	if errorReconcilerPeriod != 0 {
		after = errorReconcilerPeriod
	}
	return ctrl.Result{RequeueAfter: after}, nil
}

func retryReconcile(err error) (ctrl.Result, error) {
	return reconcile.Result{Requeue: true}, err
}
