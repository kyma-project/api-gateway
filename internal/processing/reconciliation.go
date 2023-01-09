package processing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconciliationCommand provides the processors and validation required to reconcile the API rule.
type ReconciliationCommand interface {
	// Validate performs provided APIRule validation in context of the provided client cluster
	Validate(context.Context, client.Client, *gatewayv1beta1.APIRule) ([]validation.Failure, error)

	// GetStatusForError returns ReconciliationStatus status with error in ApiRuleStatus and all other subresources status set to StatusCode
	GetStatusForError(error, validation.ResourceSelector, gatewayv1beta1.StatusCode) ReconciliationStatus

	// GetValidationStatusForFailures returns ReconciliationStatus status relevant to the encountered validation failures
	// Returns OK on handler relevant resources if supplied with an empty array
	GetValidationStatusForFailures([]validation.Failure) ReconciliationStatus

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
		log.Error(err, "Error during reconciliation")
		return cmd.GetStatusForError(err, validation.OnApiRule, gatewayv1beta1.StatusSkipped)
	}

	if len(validationFailures) > 0 {
		failuresJson, _ := json.Marshal(validationFailures)
		log.Info(fmt.Sprintf(`Validation failure {"controller": "Api", "request": "%s/%s", "failures": %s}`, apiRule.Namespace, apiRule.Name, string(failuresJson)))
		return cmd.GetValidationStatusForFailures(validationFailures)
	}

	for _, processor := range cmd.GetProcessors() {

		objectChanges, err := processor.EvaluateReconciliation(ctx, client, apiRule)
		if err != nil {
			log.Error(err, "Error during reconciliation")
			return cmd.GetStatusForError(err, validation.OnApiRule, gatewayv1beta1.StatusSkipped)
		}

		res, err := applyChanges(ctx, client, objectChanges...)
		if err != nil {
			log.Error(err, "Error during reconciliation")
			return cmd.GetStatusForError(err, res, gatewayv1beta1.StatusError)
		}
	}

	return cmd.GetValidationStatusForFailures([]validation.Failure{})
}

// applyChanges applies the given commands on the cluster
func applyChanges(ctx context.Context, client client.Client, changes ...*ObjectChange) (validation.ResourceSelector, error) {

	for _, change := range changes {
		res, err := applyChange(ctx, client, change)
		if err != nil {
			return res, err
		}
	}

	return 0, nil
}

func applyChange(ctx context.Context, client client.Client, change *ObjectChange) (validation.ResourceSelector, error) {
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

func objectToSelector(obj client.Object) validation.ResourceSelector {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	switch kind {
	case networkingv1beta1.VirtualService{}.Kind:
		return validation.OnVirtualService
	case rulev1alpha1.Rule{}.Kind:
		return validation.OnAccessRule
	case securityv1beta1.RequestAuthentication{}.Kind:
		return validation.OnRequestAuthentication
	case securityv1beta1.AuthorizationPolicy{}.Kind:
		return validation.OnAuthorizationPolicy
	default:
		return validation.OnApiRule
	}
}
