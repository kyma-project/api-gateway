package gateway

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apirulev2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
)

func (r *APIRuleReconciler) reconcileAPIRuleDeletion(ctx context.Context, log logr.Logger, apiRule *apirulev2alpha1.APIRule) (ctrl.Result, error) {
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
		if newApiRule.Spec.Gateway == nil {
			newApiRule.Spec.Gateway = ptr.To("n/a")
		}
		if newApiRule.Spec.Hosts == nil || len(newApiRule.Spec.Hosts) == 0 {
			host := apirulev2alpha1.Host("host")
			newApiRule.Spec.Hosts = []*apirulev2alpha1.Host{&host}
		}
		if newApiRule.Spec.Rules == nil {
			newApiRule.Spec.Rules = []apirulev2alpha1.Rule{{
				Methods: []apirulev2alpha1.HttpMethod{"GET"},
				Path:    "/*",
				NoAuth:  ptr.To(true),
			}}
		}
		log.Info("Deleting APIRule finalizer")
		return r.updateResourceRequeue(ctx, log, newApiRule)
	}
	return doneReconcileNoRequeue()
}
