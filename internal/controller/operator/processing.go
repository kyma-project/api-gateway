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
		//no successful reconcile recorded yet, means it's an install
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
	// a change to cr has occurred(generation increased), we set Processing
	r.log.Info("APIGateway spec changed since last successful reconcile, setting processing status",
		"APIGateway", namespacedName,
		"generation", cr.Generation,
		"observedGeneration", readyCond.ObservedGeneration,
	)
	return true

}
