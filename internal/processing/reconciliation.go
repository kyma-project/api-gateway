package processing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconciliationCommand provides the processors and validation required to reconcile the API rule.
type ReconciliationCommand interface {
	// Validate performs provided APIRule validation in context of the provided client cluster
	Validate(context.Context, client.Client, *gatewayv1beta1.APIRule) ([]validation.Failure, error)

	// GetStatusBase returns ReconciliationStatus that sets unused subresources status to nil and to gatewayv1beta1.StatusCode paramter for all the others
	GetStatusBase(gatewayv1beta1.StatusCode) ReconciliationStatus

	// GetProcessors returns the processor relevant for the reconciliation of this command.
	GetProcessors() []ReconciliationProcessor
}

// ReconciliationProcessor provides the evaluation of changes during the reconciliation of API Rule.
type ReconciliationProcessor interface {
	// EvaluateReconciliation returns the changes that needs to be applied to the cluster by comparing the desired with the actual state.
	EvaluateReconciliation(context.Context, client.Client, *gatewayv1beta1.APIRule) ([]*ObjectChange, error)
}

// Reconcile executes the reconciliation of the APIRule using the given reconciliation command.
func Reconcile(ctx context.Context, client client.Client, log *logr.Logger, cmd ReconciliationCommand, apiRule *gatewayv1beta1.APIRule) ReconciliationStatus {

	validationFailures, err := cmd.Validate(ctx, client, apiRule)
	if err != nil {
		// We set the status to skipped because it was not the validation that failed, but an error occurred during validation.
		log.Error(err, "Error during validation")
		statusBase := cmd.GetStatusBase(gatewayv1beta1.StatusSkipped)
		errorMap := map[ResourceSelector][]error{OnApiRule: {err}}
		return GetStatusForErrorMap(errorMap, statusBase)
	}

	if len(validationFailures) > 0 {
		failuresJson, _ := json.Marshal(validationFailures)
		log.Info(fmt.Sprintf(`Validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, apiRule.Namespace, apiRule.Name, string(failuresJson)))
		statusBase := cmd.GetStatusBase(gatewayv1beta1.StatusSkipped)
		return GenerateStatusFromFailures(validationFailures, statusBase)
	}

	for _, processor := range cmd.GetProcessors() {

		objectChanges, err := processor.EvaluateReconciliation(ctx, client, apiRule)
		if err != nil {
			log.Error(err, "Error during reconciliation")
			statusBase := cmd.GetStatusBase(gatewayv1beta1.StatusSkipped)
			errorMap := map[ResourceSelector][]error{OnApiRule: {err}}
			return GetStatusForErrorMap(errorMap, statusBase)
		}

		errorMap := applyChanges(ctx, client, objectChanges...)
		if len(errorMap) > 0 {
			log.Error(err, "Error during applying reconciliation")
			statusBase := cmd.GetStatusBase(gatewayv1beta1.StatusOK)
			return GetStatusForErrorMap(errorMap, statusBase)
		}
	}

	statusBase := cmd.GetStatusBase(gatewayv1beta1.StatusOK)
	return GenerateStatusFromFailures([]validation.Failure{}, statusBase)
}

// applyChanges applies the given commands on the cluster
// returns map of errors that happened for all subresources
// the map is empty if no error happened
func applyChanges(ctx context.Context, client client.Client, changes ...*ObjectChange) map[ResourceSelector][]error {
	errorMap := make(map[ResourceSelector][]error)
	for _, change := range changes {
		res, err := applyChange(ctx, client, change)
		if err != nil {
			errorMap[res] = append(errorMap[res], err)
		}
	}

	return errorMap
}

func applyChange(ctx context.Context, client client.Client, change *ObjectChange) (ResourceSelector, error) {
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

func objectToSelector(obj client.Object) ResourceSelector {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	switch kind {
	case OnVirtualService.String():
		return OnVirtualService
	case OnAccessRule.String():
		return OnAccessRule
	case OnRequestAuthentication.String():
		return OnRequestAuthentication
	case OnAuthorizationPolicy.String():
		return OnAuthorizationPolicy
	default:
		return OnApiRule
	}
}
