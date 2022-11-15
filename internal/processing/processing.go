package processing

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

var (
	//OwnerLabel .
	OwnerLabel = fmt.Sprintf("%s.%s", "apirule", gatewayv1beta1.GroupVersion.String())
	//OwnerLabelv1alpha1 .
	OwnerLabelv1alpha1 = fmt.Sprintf("%s.%s", "apirule", gatewayv1alpha1.GroupVersion.String())
)

// Reconcile executes the reconciliation of the APIRule using the given processors.
func Reconcile(apiRule *gatewayv1beta1.APIRule, processors []ReconciliationProcessor) (gatewayv1beta1.StatusCode, error) {
	// TODO: There are multiple ways to improve this. We could do it async, but the added complexity might not be worth it.
	//  And we can return the status for each processor/object kind to the caller.
	for _, processor := range processors {

		status, err := processor.Reconcile(apiRule)

		if status != gatewayv1beta1.StatusOK {
			return status, err
		}
	}

	return gatewayv1beta1.StatusOK, nil
}

// applyDiff applies the given objects state on the cluster
func applyDiff(ctx context.Context, client client.Client, objectsToApply ...*ObjToPatch) error {

	for _, object := range objectsToApply {
		err := applyObjDiff(ctx, client, object)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyObjDiff(ctx context.Context, client client.Client, objToPatch *ObjToPatch) error {
	var err error

	switch objToPatch.action {
	case "create":
		err = client.Create(ctx, objToPatch.obj)
	case "update":
		err = client.Update(ctx, objToPatch.obj)
	case "delete":
		err = client.Delete(ctx, objToPatch.obj)
	default:
		err = fmt.Errorf("apply action %s is not supported", objToPatch.action)
	}

	if err != nil {
		return err
	}

	return nil
}
