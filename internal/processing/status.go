package processing

import (
	"fmt"

	"github.com/go-logr/logr"
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
)

type ReconciliationStatus struct {
	ApiRuleStatus               *gatewayv1beta1.APIRuleResourceStatus
	VirtualServiceStatus        *gatewayv1beta1.APIRuleResourceStatus
	AccessRuleStatus            *gatewayv1beta1.APIRuleResourceStatus
	RequestAuthenticationStatus *gatewayv1beta1.APIRuleResourceStatus
	AuthorizationPolicyStatus   *gatewayv1beta1.APIRuleResourceStatus
}

func getStatus(apiStatus *gatewayv1beta1.APIRuleResourceStatus, statusCode gatewayv1beta1.StatusCode) ReconciliationStatus {
	return ReconciliationStatus{
		ApiRuleStatus: apiStatus,
		VirtualServiceStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		}, AccessRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		}, RequestAuthenticationStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		}, AuthorizationPolicyStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		},
	}
}

func getOkStatus() ReconciliationStatus {
	apiRuleStatus := &gatewayv1beta1.APIRuleResourceStatus{
		Code: gatewayv1beta1.StatusOK,
	}
	return getStatus(apiRuleStatus, gatewayv1beta1.StatusOK)
}

// GetStatusForError creates a status with APIRule status in error condition. Accepts an auxiliary status code that is used to report VirtualService and AccessRule status.
func GetStatusForError(log *logr.Logger, err error, statusCode gatewayv1beta1.StatusCode) ReconciliationStatus {
	log.Error(err, "Error during reconciliation")
	return ReconciliationStatus{
		ApiRuleStatus: generateErrorStatus(err),
		VirtualServiceStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		}, AccessRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		}, RequestAuthenticationStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		}, AuthorizationPolicyStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: statusCode,
		},
	}
}

func generateErrorStatus(err error) *gatewayv1beta1.APIRuleResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, err.Error())
}

func GetFailedValidationStatus(failures []validation.Failure) ReconciliationStatus {
	apiRuleStatus := generateValidationStatus(failures)
	return getStatus(apiRuleStatus, gatewayv1beta1.StatusSkipped)
}

func generateValidationStatus(failures []validation.Failure) *gatewayv1beta1.APIRuleResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, generateValidationDescription(failures))
}

func generateValidationDescription(failures []validation.Failure) string {
	var description string

	if len(failures) == 1 {
		description = "Validation error: "
		description += fmt.Sprintf("Attribute \"%s\": %s", failures[0].AttributePath, failures[0].Message)
	} else {
		const maxEntries = 3
		description = "Multiple validation errors: "
		for i := 0; i < len(failures) && i < maxEntries; i++ {
			description += fmt.Sprintf("\nAttribute \"%s\": %s", failures[i].AttributePath, failures[i].Message)
		}
		if len(failures) > maxEntries {
			description += fmt.Sprintf("\n%d more error(s)...", len(failures)-maxEntries)
		}
	}

	return description
}

func toStatus(c gatewayv1beta1.StatusCode, desc string) *gatewayv1beta1.APIRuleResourceStatus {
	return &gatewayv1beta1.APIRuleResourceStatus{
		Code:        c,
		Description: desc,
	}
}
