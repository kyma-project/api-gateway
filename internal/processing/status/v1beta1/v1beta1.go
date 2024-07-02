package v1beta1

import (
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/versions"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	"github.com/kyma-project/api-gateway/internal/validation"
)

type ReconciliationV1beta1Status struct {
	ApiRuleStatus               *gatewayv1beta1.APIRuleResourceStatus
	VirtualServiceStatus        *gatewayv1beta1.APIRuleResourceStatus
	AccessRuleStatus            *gatewayv1beta1.APIRuleResourceStatus
	RequestAuthenticationStatus *gatewayv1beta1.APIRuleResourceStatus
	AuthorizationPolicyStatus   *gatewayv1beta1.APIRuleResourceStatus
}

func (s ReconciliationV1beta1Status) HasError() bool {
	if s.ApiRuleStatus != nil && s.ApiRuleStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if s.VirtualServiceStatus != nil && s.VirtualServiceStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if s.AccessRuleStatus != nil && s.AccessRuleStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if s.AuthorizationPolicyStatus != nil && s.AuthorizationPolicyStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	if s.RequestAuthenticationStatus != nil && s.RequestAuthenticationStatus.Code == gatewayv1beta1.StatusError {
		return true
	}
	return false
}

func (s ReconciliationV1beta1Status) GetStatusForErrorMap(errorMap map[status.ResourceSelector][]error) status.ReconciliationStatusVisitor {
	for key, val := range errorMap {
		switch key {
		case status.OnApiRule:
			s.ApiRuleStatus = generateStatusFromErrors(val)
		case status.OnVirtualService:
			s.VirtualServiceStatus = generateStatusFromErrors(val)
		case status.OnAccessRule:
			s.AccessRuleStatus = generateStatusFromErrors(val)
		case status.OnAuthorizationPolicy:
			s.AuthorizationPolicyStatus = generateStatusFromErrors(val)
		case status.OnRequestAuthentication:
			s.RequestAuthenticationStatus = generateStatusFromErrors(val)
		}

		if key != status.OnApiRule {
			if s.ApiRuleStatus == nil || s.ApiRuleStatus.Code == gatewayv1beta1.StatusOK {
				s.ApiRuleStatus = &gatewayv1beta1.APIRuleResourceStatus{
					Code:        gatewayv1beta1.StatusError,
					Description: fmt.Sprintf("Error has happened on subresource %s", key),
				}
			} else {
				s.ApiRuleStatus.Code = gatewayv1beta1.StatusError
				s.ApiRuleStatus.Description += fmt.Sprintf("\nError has happened on subresource %s", key)
			}
		}
	}

	return s
}

func (s ReconciliationV1beta1Status) GenerateStatusFromFailures(failures []validation.Failure) status.ReconciliationStatusVisitor {
	if len(failures) == 0 {
		return s
	}

	s.ApiRuleStatus = generateValidationStatus(failures)
	return s
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

func (s ReconciliationV1beta1Status) VisitStatus(status status.Status) error {
	if status.ApiRuleStatusVersion() != versions.V1beta1 {
		return fmt.Errorf("v1beta1 status visitor cannot handle status of version %s", status.ApiRuleStatusVersion())
	}

	v1beta1Status := status.(*gatewayv1beta1.APIRuleStatus)
	v1beta1Status.APIRuleStatus = s.ApiRuleStatus
	v1beta1Status.VirtualServiceStatus = s.VirtualServiceStatus
	v1beta1Status.AccessRuleStatus = s.AccessRuleStatus
	v1beta1Status.RequestAuthenticationStatus = s.RequestAuthenticationStatus
	v1beta1Status.AuthorizationPolicyStatus = s.AuthorizationPolicyStatus

	return nil
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
