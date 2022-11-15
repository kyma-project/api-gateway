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
func Reconcile(client client.Client, ctx context.Context, processors []ReconciliationProcessor, apiRule *gatewayv1beta1.APIRule) (gatewayv1beta1.StatusCode, error) {
	// TODO: There are multiple ways to improve this. We could do it async, but the added complexity might not be worth it.
	//  And we can return the status for each processor/object kind to the caller.
	for _, processor := range processors {

		cmds, status, err := processor.evaluateReconciliation(apiRule)
		if status != gatewayv1beta1.StatusOK {
			return status, err
		}

		// TODO: Should we apply all changes after all commands are evaluated?
		err = applyReconciliation(ctx, client, cmds...)
		if err != nil {
			// TODO: Old comment was:
			//  "We don't know exactly which object(s) are not updated properly. The safest approach is to assume nothing is correct and just use `StatusError`."
			//  Can we improve the Status here?
			return gatewayv1beta1.StatusError, err
		}
	}

	return gatewayv1beta1.StatusOK, nil
}

// applyReconciliation applies the given commands on the cluster
func applyReconciliation(ctx context.Context, client client.Client, cmds ...*ReconciliationCommand) error {

	for _, object := range cmds {
		err := applyCommand(ctx, client, object)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyCommand(ctx context.Context, client client.Client, cmd *ReconciliationCommand) error {
	var err error

	switch cmd.action {
	case "create":
		err = client.Create(ctx, cmd.obj)
	case "update":
		err = client.Update(ctx, cmd.obj)
	case "delete":
		err = client.Delete(ctx, cmd.obj)
	default:
		err = fmt.Errorf("apply action %s is not supported", cmd.action)
	}

	if err != nil {
		return err
	}

	return nil
}

func GetReconciliationProcessors(config ReconciliationConfig, apiRule *gatewayv1beta1.APIRule) []ReconciliationProcessor {
	// TODO This apiRule check is just a dummy for real checks
	if apiRule != nil {
		return []ReconciliationProcessor{NewVirtualServiceProcessor(config), NewAccessRuleProcessor(config)}
	} else {
		return []ReconciliationProcessor{}
	}
}
