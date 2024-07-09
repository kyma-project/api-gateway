package gateway

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *APIRuleReconciler) reconcileAPIRuleDeletion(ctx context.Context, apiRule *gatewayv1beta1.APIRule) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(apiRule, apiGatewayFinalizer) {
		// finalizer is present on APIRule, so all subresources need to be deleted
		if err := processing.DeleteAPIRuleSubresources(r.Client, ctx, *apiRule); err != nil {
			r.Log.Error(err, "Error happened during deletion of APIRule subresources")
			// if removing subresources ends in error, return with retry
			// so that it can be retried
			return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
		}

		// remove finalizer so the reconciliation can proceed
		controllerutil.RemoveFinalizer(apiRule, apiGatewayFinalizer)
		// workaround for when APIRule was deleted using v2alpha1 version and it got trimmed spec
		if apiRule.Spec.Gateway == nil {
			apiRule.Spec.Gateway = ptr.To("n/a")
		}
		if apiRule.Spec.Host == nil {
			apiRule.Spec.Host = ptr.To("host")
		}
		if apiRule.Spec.Rules == nil {
			apiRule.Spec.Rules = []gatewayv1beta1.Rule{{
				Methods: []gatewayv1beta1.HttpMethod{"GET"},
				Path:    "/*",
				AccessStrategies: []*gatewayv1beta1.Authenticator{
					{
						Handler: &gatewayv1beta1.Handler{
							Name: "noop",
						},
					},
				},
			}}
		}
		if err := r.Update(ctx, apiRule); err != nil {
			r.Log.Error(err, "Error happened during finalizer removal")
			return doneReconcileErrorRequeue(r.OnErrorReconcilePeriod)
		}
	}
	return doneReconcileNoRequeue()
}
