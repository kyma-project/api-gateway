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

func generateErrorStatus(err error) *gatewayv1beta1.ResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, err.Error())
}

func GenerateStatusFromFailures(failures []validation.Failure) *gatewayv1beta1.ResourceStatus {
	if len(failures) == 0 {
		return &gatewayv1beta1.ResourceStatus{Code: gatewayv1beta1.StatusOK}
	}

	return &gatewayv1beta1.ResourceStatus{Code: gatewayv1beta1.StatusError, Description: generateValidationDescription(failures)}
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
