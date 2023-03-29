package processing

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

type ReconciliationStatus struct {
	ApiRuleStatus               *gatewayv1beta1.APIRuleResourceStatus
	VirtualServiceStatus        *gatewayv1beta1.APIRuleResourceStatus
	AccessRuleStatus            *gatewayv1beta1.APIRuleResourceStatus
	RequestAuthenticationStatus *gatewayv1beta1.APIRuleResourceStatus
	AuthorizationPolicyStatus   *gatewayv1beta1.APIRuleResourceStatus
}

func (status ReconciliationStatus) HasError() bool {
	if status.ApiRuleStatus != nil && status.ApiRuleStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if status.VirtualServiceStatus != nil && status.VirtualServiceStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if status.AccessRuleStatus != nil && status.AccessRuleStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if status.AuthorizationPolicyStatus != nil && status.AuthorizationPolicyStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if status.RequestAuthenticationStatus != nil && status.RequestAuthenticationStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	return false
}

type ResourceSelector int

const (
	OnApiRule ResourceSelector = iota
	OnVirtualService
	OnAccessRule
	OnAuthorizationPolicy
	OnRequestAuthentication
)

func (r ResourceSelector) String() string {
	switch r {
	case OnVirtualService:
		return "VirtualService"
	case OnAccessRule:
		return "Rule"
	case OnRequestAuthentication:
		return "RequestAuthentication"
	case OnAuthorizationPolicy:
		return "AuthorizationPolicy"
	default:
		// If no Kind is resolved from the resource (e.g. subresource CRD is missing)
		return "APIRule"
	}
}

func generateStatusFromErrors(errors []error) *gatewayv1beta1.APIRuleResourceStatus {
	status := &gatewayv1beta1.APIRuleResourceStatus{}
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

func GetStatusForErrorMap(errorMap map[ResourceSelector][]error, statusBase ReconciliationStatus) ReconciliationStatus {
	for key, val := range errorMap {
		switch key {
		case OnApiRule:
			statusBase.ApiRuleStatus = generateStatusFromErrors(val)
		case OnVirtualService:
			statusBase.VirtualServiceStatus = generateStatusFromErrors(val)
		case OnAccessRule:
			statusBase.AccessRuleStatus = generateStatusFromErrors(val)
		case OnAuthorizationPolicy:
			statusBase.AuthorizationPolicyStatus = generateStatusFromErrors(val)
		case OnRequestAuthentication:
			statusBase.RequestAuthenticationStatus = generateStatusFromErrors(val)
		}

		if key != OnApiRule {
			if statusBase.ApiRuleStatus == nil || statusBase.ApiRuleStatus.Code == gatewayv1beta1.StatusOK {
				statusBase.ApiRuleStatus = &gatewayv1beta1.APIRuleResourceStatus{
					Code:        gatewayv1beta1.StatusError,
					Description: fmt.Sprintf("Error has happened on subresource %s", key),
				}
			} else {
				statusBase.ApiRuleStatus.Code = gatewayv1beta1.StatusError
				statusBase.ApiRuleStatus.Description += fmt.Sprintf("\nError has happened on subresource %s", key)
			}
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
