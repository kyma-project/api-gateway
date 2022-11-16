package processing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/controllers"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconciliationCommand interface {
	validate(*gatewayv1beta1.APIRule) ([]validation.Failure, error)
	getLogger() logr.Logger
	getProcessors() []ReconciliationProcessor
	getContext() context.Context
	getClient() client.Client
}

// TODO: We should return a more meaningful structure than two status codes of different types

// Reconcile executes the reconciliation of the APIRule using the given reconciliation command.
func Reconcile(cmd ReconciliationCommand, apiRule *gatewayv1beta1.APIRule) (gatewayv1beta1.StatusCode, *gatewayv1beta1.APIRuleResourceStatus, error) {

	validationFailures, err := cmd.validate(apiRule)

	if err != nil {
		// We set the status to skipped because it was not the validation that failed, but an error occurred during validation.
		return gatewayv1beta1.StatusSkipped, nil, err
	}

	if len(validationFailures) > 0 {
		failuresJson, _ := json.Marshal(validationFailures)
		cmd.getLogger().Info(fmt.Sprintf(`Validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, apiRule.Namespace, apiRule.Name, string(failuresJson)))
		return gatewayv1beta1.StatusSkipped, controllers.GenerateValidationStatus(validationFailures), nil
	}

	// TODO: There are multiple ways to improve this. We could do it async, but the added complexity might not be worth it.
	//  And we can return the status for each processor/object kind to the caller.
	for _, processor := range cmd.getProcessors() {

		objectChanges, status, err := processor.EvaluateReconciliation(apiRule)
		if status != gatewayv1beta1.StatusOK {
			return status, nil, err
		}

		// TODO: Should we apply all changes after all commands are evaluated?
		err = applyChanges(cmd.getContext(), cmd.getClient(), objectChanges...)
		if err != nil {
			// TODO: Old comment was:
			//  "We don't know exactly which object(s) are not updated properly. The safest approach is to assume nothing is correct and just use `StatusError`."
			//  Can we improve the Status here?
			return gatewayv1beta1.StatusError, nil, err
		}
	}

	return gatewayv1beta1.StatusOK, nil, nil
}

// applyChanges applies the given commands on the cluster
func applyChanges(ctx context.Context, client client.Client, changes ...*ObjectChange) error {

	for _, change := range changes {
		err := applyChange(ctx, client, change)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyChange(ctx context.Context, client client.Client, change *ObjectChange) error {
	var err error

	switch change.action {
	case "create":
		err = client.Create(ctx, change.obj)
	case "update":
		err = client.Update(ctx, change.obj)
	case "delete":
		err = client.Delete(ctx, change.obj)
	default:
		err = fmt.Errorf("apply action %s is not supported", change.action)
	}

	if err != nil {
		return err
	}

	return nil
}
