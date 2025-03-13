package gateway

import (
	"context"
	"github.com/go-logr/logr"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *APIRuleReconciler) reconcileAPIRuleDeletion(ctx context.Context, log logr.Logger, apiRule *gatewayv2alpha1.APIRule) (ctrl.Result, error) {
	newApiRule := apiRule.DeepCopy()
	if controllerutil.ContainsFinalizer(apiRule, apiGatewayFinalizer) {
		// finalizer is present on APIRule, so all subresources need to be deleted
		if err := processing.DeleteAPIRuleSubresources(r.Client, ctx, *apiRule); err != nil {
			log.Error(err, "Error happened during deletion of APIRule subresources")
			// if removing subresources ends in error, return with retry
			// so that it can be retried
			return doneReconcileErrorRequeue(err, r.OnErrorReconcilePeriod)
		}

		// remove finalizer so the reconciliation can proceed
		controllerutil.RemoveFinalizer(newApiRule, apiGatewayFinalizer)
		// workaround for when APIRule was deleted using v2alpha1 version
		// and if it got trimmed spec
		//if newApiRule.Spec.Gateway == nil {
		//	newApiRule.Spec.Gateway = ptr.To("n/a")
		//}
		//if newApiRule.Spec.Hosts == nil || len(newApiRule.Spec.Hosts) == 0 {
		//	host := gatewayv2alpha1.Host("host")
		//	newApiRule.Spec.Hosts = []*gatewayv2alpha1.Host{&host}
		//}
		//if newApiRule.Spec.Rules == nil {
		//	newApiRule.Spec.Rules = []gatewayv1beta1.Rule{{
		//		Methods: []gatewayv1beta1.HttpMethod{"GET"},
		//		Path:    "/*",
		//		AccessStrategies: []*gatewayv1beta1.Authenticator{
		//			{
		//				Handler: &gatewayv1beta1.Handler{
		//					Name: "noop",
		//				},
		//			},
		//		},
		//	}}
		//}
		log.Info("Deleting APIRule finalizer")
		return r.updateResourceRequeue(ctx, log, newApiRule)
	}
	return doneReconcileNoRequeue()
}
