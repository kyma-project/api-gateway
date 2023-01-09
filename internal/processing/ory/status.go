package ory

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"github.com/kyma-incubator/api-gateway/internal/validation"
)

func (r Reconciliation) GetStatusForError(err error, selector validation.ResourceSelector, statusCode gatewayv1beta1.StatusCode) (status processing.ReconciliationStatus) {
	status = processing.ReconciliationStatus{
		ApiRuleStatus: &gatewayv1beta1.ResourceStatus{
			Code: statusCode,
		}, VirtualServiceStatus: &gatewayv1beta1.ResourceStatus{
			Code: statusCode,
		}, AccessRuleStatus: &gatewayv1beta1.ResourceStatus{
			Code: statusCode,
		},
	}

	switch selector{
	case validation.OnApiRule:
		status.ApiRuleStatus = generateErrorStatus(err)
	case validation.OnVirtualService:
		status.VirtualServiceStatus = generateErrorStatus(err)
	case validation.OnAccessRule:
		status.AccessRuleStatus = generateErrorStatus(err)
	}

	return status
}

func generateErrorStatus(err error) *gatewayv1beta1.ResourceStatus {
	return toStatus(gatewayv1beta1.StatusError, err.Error())
}

func toStatus(c gatewayv1beta1.StatusCode, desc string) *gatewayv1beta1.ResourceStatus {
	return &gatewayv1beta1.ResourceStatus{
		Code:        c,
		Description: desc,
	}
}

func (r Reconciliation) GetValidationStatusForFailures(failures []validation.Failure) (status processing.ReconciliationStatus) {
	failuresMap := make(map[validation.ResourceSelector][]validation.Failure)
	for _, failure := range failures {
		failuresMap[failure.OnResource] = append(failuresMap[failure.OnResource], failure)
	}

	status.ApiRuleStatus = processing.GenerateStatusFromFailures(failuresMap[validation.OnApiRule])
	status.VirtualServiceStatus = processing.GenerateStatusFromFailures(failuresMap[validation.OnApiRule])
	status.AccessRuleStatus = processing.GenerateStatusFromFailures(failuresMap[validation.OnApiRule])

	// If the status for AP and RA is OK, the field is set to nil
	// If an error has happened, the error is caused by using invalid handler (in this case Ory) for the configuration
	authorizationPolicyStatus := processing.GenerateStatusFromFailures(failuresMap[validation.OnApiRule])
	if authorizationPolicyStatus.Code != gatewayv1beta1.StatusOK {
		status.AuthorizationPolicyStatus = authorizationPolicyStatus
	}

	requestAuthenticationStatus := processing.GenerateStatusFromFailures(failuresMap[validation.OnApiRule])
	if requestAuthenticationStatus.Code != gatewayv1beta1.StatusOK {
		status.RequestAuthenticationStatus = requestAuthenticationStatus
	}

	return status
}
