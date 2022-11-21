package processing

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconciliationCommand interface {
	Validate(*gatewayv1beta1.APIRule) ([]validation.Failure, error)
	GetLogger() logr.Logger
	GetProcessors() []ReconciliationProcessor
	GetContext() context.Context
	GetClient() client.Client
}

type ReconciliationProcessor interface {
	EvaluateReconciliation(*gatewayv1beta1.APIRule) ([]*ObjectChange, error)
}

// Reconcile executes the reconciliation of the APIRule using the given reconciliation command.
func Reconcile(cmd ReconciliationCommand, apiRule *gatewayv1beta1.APIRule) ReconciliationStatus {

	validationFailures, err := cmd.Validate(apiRule)
	if err != nil {
		// We set the status to skipped because it was not the validation that failed, but an error occurred during validation.
		return GetStatusForError(cmd.GetLogger(), err, gatewayv1beta1.StatusSkipped)
	}

	if len(validationFailures) > 0 {
		failuresJson, _ := json.Marshal(validationFailures)
		cmd.GetLogger().Info(fmt.Sprintf(`Validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, apiRule.Namespace, apiRule.Name, string(failuresJson)))
		return getFailedValidationStatus(validationFailures)
	}

	for _, processor := range cmd.GetProcessors() {

		objectChanges, err := processor.EvaluateReconciliation(apiRule)
		if err != nil {
			return GetStatusForError(cmd.GetLogger(), err, gatewayv1beta1.StatusSkipped)
		}

		err = applyChanges(cmd.GetContext(), cmd.GetClient(), objectChanges...)
		if err != nil {
			//  "We don't know exactly which object(s) are not updated properly. The safest approach is to assume nothing is correct and just use `StatusError`."
			return GetStatusForError(cmd.GetLogger(), err, gatewayv1beta1.StatusError)
		}
	}

	return getOkStatus()
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

	switch change.Action {
	case "create":
		err = client.Create(ctx, change.Obj)
	case "update":
		err = client.Update(ctx, change.Obj)
	case "delete":
		err = client.Delete(ctx, change.Obj)
	default:
		err = fmt.Errorf("apply action %s is not supported", change.Action)
	}

	if err != nil {
		return err
	}

	return nil
}
