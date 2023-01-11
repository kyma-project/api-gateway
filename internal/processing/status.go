package processing

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
)

type ReconciliationStatus struct {
	ApiRuleStatus               *gatewayv1beta1.ResourceStatus
	VirtualServiceStatus        *gatewayv1beta1.ResourceStatus
	AccessRuleStatus            *gatewayv1beta1.ResourceStatus
	RequestAuthenticationStatus *gatewayv1beta1.ResourceStatus
	AuthorizationPolicyStatus   *gatewayv1beta1.ResourceStatus
}

func generateStatusFromErrors(errors []error) *gatewayv1beta1.ResourceStatus {
	status := &gatewayv1beta1.ResourceStatus{}
	if len(errors) == 0 {
		status.Code = gatewayv1beta1.StatusOK
		return status
	}
	status.Code = gatewayv1beta1.StatusError
	status.Description = errors[0].Error()
	for _, err := range errors[1:] {
		status.Description = fmt.Sprintf("%s\n%s", status.Description, err.Error())
	}
	return status
}

func GetStatusForErrorMap(errorMap map[validation.ResourceSelector][]error, statusBase ReconciliationStatus) ReconciliationStatus {
	for key, val := range errorMap {
		switch key {
		case validation.OnApiRule:
			statusBase.ApiRuleStatus = generateStatusFromErrors(val)
		case validation.OnVirtualService:
			statusBase.VirtualServiceStatus = generateStatusFromErrors(val)
		case validation.OnAccessRule:
			statusBase.AccessRuleStatus = generateStatusFromErrors(val)
		case validation.OnAuthorizationPolicy:
			statusBase.AuthorizationPolicyStatus = generateStatusFromErrors(val)
		case validation.OnRequestAuthentication:
			statusBase.RequestAuthenticationStatus = generateStatusFromErrors(val)
		}
	}

	return statusBase
}

func GenerateStatusFromFailures(failures []validation.Failure, statusBase ReconciliationStatus) ReconciliationStatus {
	if len(failures) == 0 {
		return statusBase
	}

	statusBase.ApiRuleStatus = generateValidationStatus(failures)
	return statusBase
}

func generateValidationStatus(failures []validation.Failure) *gatewayv1beta1.ResourceStatus {
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

func toStatus(c gatewayv1beta1.StatusCode, desc string) *gatewayv1beta1.ResourceStatus {
	return &gatewayv1beta1.ResourceStatus{
		Code:        c,
		Description: desc,
	}
}
