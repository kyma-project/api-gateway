package v2alpha1

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	v1beta1Status "github.com/kyma-project/api-gateway/internal/processing/status"
)

func (r Reconciliation) GetStatusBase(statusCode string) status.ReconciliationStatus {
	return StatusBase(statusCode)
}

func StatusBase(statusCode string) status.ReconciliationStatus {
	return v1beta1Status.ReconciliationV1beta1Status{
		ApiRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusCode(statusCode),
		},
		VirtualServiceStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusCode(statusCode),
		},
		AuthorizationPolicyStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusCode(statusCode),
		},
		RequestAuthenticationStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusCode(statusCode),
		},
	}
}
