package operator

import (
	"context"

	operatorv1alpha1 "github.com/kyma-project/api-gateway/apis/operator/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
)

func (r *APIGatewayReconciler) shouldSetProcessing(ctx context.Context, namespacedName types.NamespacedName) bool {
	cr := operatorv1alpha1.APIGateway{}
	if err := r.Get(ctx, namespacedName, &cr); err != nil {
		r.log.Error(err, "Failed to get APIGateway resource", "APIGateway", namespacedName)
		return false
	}

	if cr.IsInDeletion() {
		r.log.Info("APIGateway is being deleted, skipping processing status update", "APIGateway", namespacedName)
		return false
	}

	readyCond := meta.FindStatusCondition(cr.Status.Conditions, "Ready")
	if readyCond == nil {
		r.log.Info("APIGateway has no Ready condition yet, setting processing status", "APIGateway", namespacedName)
		return true
	}

	if cr.Generation <= readyCond.ObservedGeneration {
		r.log.Info("APIGateway resource has not changed since last successful reconcile, skipping processing status update",
			"APIGateway", namespacedName,
			"generation", cr.Generation,
			"observedGeneration", readyCond.ObservedGeneration,
		)
		return false
	}

	//spec changed, only set Processing if enableKymaGateway is being disabled, as that causes downtime
	if isKymaGatewayBeingDisabled(&cr) {
		r.log.Info("enableKymaGateway is being disabled, setting processing status",
			"APIGateway", namespacedName,
		)
		return true
	}

	r.log.Info("APIGateway spec changed but no downtime-causing change detected, skipping processing status update",
		"APIGateway", namespacedName,
	)
	return false
}

// isKymaGatewayBeingDisabled returns true when the current spec disables the Kyma Gateway
// and the last successfully reconciled state had it enabled (tracked via annotation).
// Disabling the gateway deletes the Istio Gateway resource and causes downtime
func isKymaGatewayBeingDisabled(cr *operatorv1alpha1.APIGateway) bool {
	currentlyDisabled := cr.Spec.EnableKymaGateway == nil || !*cr.Spec.EnableKymaGateway
	if !currentlyDisabled {
		return false
	}
	lastApplied, ok := cr.GetAnnotations()[operatorv1alpha1.LastAppliedEnableKymaGatewayAnnotation]
	return ok && lastApplied == "true"
}
