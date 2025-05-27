package ory

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	v1beta1Status "github.com/kyma-project/api-gateway/internal/processing/status"
)

func (r Reconciliation) GetStatusBase(statusCode string) v1beta1Status.ReconciliationStatus {
	return v1beta1Status.ReconciliationV1beta1Status{
		APIRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusCode(statusCode),
		},
		VirtualServiceStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusCode(statusCode),
		},
		AccessRuleStatus: &gatewayv1beta1.APIRuleResourceStatus{
			Code: gatewayv1beta1.StatusCode(statusCode),
		},
	}
}
