package istio

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
)

func (r Reconciliation) GetStatusBase(statusCode gatewayv1beta1.StatusCode) processing.ReconciliationStatus {
	return processing.ReconciliationStatus{
		ApiRuleStatus: &gatewayv1beta1.ResourceStatus{
			Code: statusCode,
		},
		VirtualServiceStatus: &gatewayv1beta1.ResourceStatus{
			Code: statusCode,
		},
		AuthorizationPolicyStatus: &gatewayv1beta1.ResourceStatus{
			Code: statusCode,
		},
		RequestAuthenticationStatus: &gatewayv1beta1.ResourceStatus{
			Code: statusCode,
		},
	}

}
