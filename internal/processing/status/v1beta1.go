package status

import (
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/apis/gateway/versions"
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

func (s ReconciliationV1beta1Status) GetStatusForErrorMap(errorMap map[ResourceSelector][]error) ReconciliationStatus {
	for key, val := range errorMap {
		switch key {
		case OnApiRule:
			s.ApiRuleStatus = generateStatusFromErrors(val)
		case OnVirtualService:
			s.VirtualServiceStatus = generateStatusFromErrors(val)
		case OnAccessRule:
			s.AccessRuleStatus = generateStatusFromErrors(val)
		case OnAuthorizationPolicy:
			s.AuthorizationPolicyStatus = generateStatusFromErrors(val)
		case OnRequestAuthentication:
			s.RequestAuthenticationStatus = generateStatusFromErrors(val)
		}

		if key != OnApiRule {
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

func (s ReconciliationV1beta1Status) GenerateStatusFromFailures(failures []validation.Failure) ReconciliationStatus {
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

func (s ReconciliationV1beta1Status) UpdateStatus(status *gatewayv1beta1.APIRuleStatus) error {
	if status.ApiRuleStatusVersion() != versions.V1beta1 {
		return fmt.Errorf("v1beta1 status handler cannot handle status of version %s", status.ApiRuleStatusVersion())
	}

	status.APIRuleStatus = s.ApiRuleStatus
	status.VirtualServiceStatus = s.VirtualServiceStatus
	status.AccessRuleStatus = s.AccessRuleStatus
	status.RequestAuthenticationStatus = s.RequestAuthenticationStatus
	status.AuthorizationPolicyStatus = s.AuthorizationPolicyStatus

	return nil
}

func generateValidationStatus(failures []validation.Failure) *gatewayv1beta1.APIRuleResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, generateValidationDescription(failures))
}

func toStatus(c gatewayv1beta1.StatusCode, desc string) *gatewayv1beta1.APIRuleResourceStatus {
	return &gatewayv1beta1.APIRuleResourceStatus{
		Code:        c,
		Description: desc,
	}
}
