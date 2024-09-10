package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	"github.com/kyma-project/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconciliationCommand provides the processors and validation required to reconcile the API rule.
type ReconciliationCommand interface {
	// Validate performs provided APIRule validation in context of the provided client cluster
	Validate(context.Context, client.Client) ([]validation.Failure, error)

	// GetStatusBase returns ReconciliationV1beta1Status that sets unused subresources status to nil and to gatewayv1beta1.StatusCode paramter for all the others
	GetStatusBase(string) status.ReconciliationStatus

	// GetProcessors returns the processor relevant for the reconciliation of this command.
	GetProcessors() []ReconciliationProcessor
}

// ReconciliationProcessor provides the evaluation of changes during the reconciliation of API Rule.
type ReconciliationProcessor interface {
	// EvaluateReconciliation returns the changes that needs to be applied to the cluster by comparing the desired with the actual state.
	EvaluateReconciliation(context.Context, client.Client) ([]*ObjectChange, error)
}

// Reconcile executes the reconciliation of the APIRule using the given reconciliation command.
func Reconcile(ctx context.Context, client client.Client, log *logr.Logger, cmd ReconciliationCommand) status.ReconciliationStatus {
	l := log.WithValues("controller", "APIRule", "version", gatewayv1beta1.GroupVersion.String())

	validationFailures, err := cmd.Validate(ctx, client)
	if err != nil {
		// We set the status to skipped because it was not the validation that failed, but an error occurred during validation.
		l.Error(err, "Error during validation")
		statusBase := cmd.GetStatusBase(string(gatewayv1beta1.StatusSkipped))
		errorMap := map[status.ResourceSelector][]error{status.OnApiRule: {err}}
		return statusBase.GetStatusForErrorMap(errorMap)
	}

	if len(validationFailures) > 0 {
		failuresJson, _ := json.Marshal(validationFailures)
		l.Error(errors.New("validation failure"), "Validation failure", "failure", string(failuresJson))
		statusBase := cmd.GetStatusBase(string(gatewayv1beta1.StatusSkipped))
		return statusBase.GenerateStatusFromFailures(validationFailures)
	}

	for _, processor := range cmd.GetProcessors() {

		objectChanges, err := processor.EvaluateReconciliation(ctx, client)
		if err != nil {
			l.Error(err, "Error during reconciliation")
			statusBase := cmd.GetStatusBase(string(gatewayv1beta1.StatusSkipped))
			errorMap := map[status.ResourceSelector][]error{status.OnApiRule: {err}}
			return statusBase.GetStatusForErrorMap(errorMap)
		}

		errorMap := applyChanges(ctx, client, objectChanges...)
		if len(errorMap) > 0 {
			l.Error(err, "Error during applying reconciliation")
			statusBase := cmd.GetStatusBase(string(gatewayv1beta1.StatusOK))
			return statusBase.GetStatusForErrorMap(errorMap)
		}
	}

	statusBase := cmd.GetStatusBase(string(gatewayv1beta1.StatusOK))
	return statusBase.GenerateStatusFromFailures(nil)
}

// applyChanges applies the given commands on the cluster
// returns map of errors that happened for all subresources
// the map is empty if no error happened
func applyChanges(ctx context.Context, client client.Client, changes ...*ObjectChange) map[status.ResourceSelector][]error {
	errorMap := make(map[status.ResourceSelector][]error)
	for _, change := range changes {
		res, err := applyChange(ctx, client, change)
		if err != nil {
			errorMap[res] = append(errorMap[res], err)
		}
	}

	return errorMap
}

func applyChange(ctx context.Context, client client.Client, change *ObjectChange) (status.ResourceSelector, error) {
	var err error
	switch change.Action {
	case create:
		err = client.Create(ctx, change.Obj)
	case update:
		err = client.Update(ctx, change.Obj)
	case delete:
		err = client.Delete(ctx, change.Obj)
	default:
		err = fmt.Errorf("apply action %s is not supported", change.Action)
	}

	if err != nil {
		return objectToSelector(change.Obj), err
	}

	return objectToSelector(change.Obj), nil
}

func objectToSelector(obj client.Object) status.ResourceSelector {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	switch kind {
	case status.OnVirtualService.String():
		return status.OnVirtualService
	case status.OnAccessRule.String():
		return status.OnAccessRule
	case status.OnRequestAuthentication.String():
		return status.OnRequestAuthentication
	case status.OnAuthorizationPolicy.String():
		return status.OnAuthorizationPolicy
	default:
		return status.OnApiRule
	}
}
