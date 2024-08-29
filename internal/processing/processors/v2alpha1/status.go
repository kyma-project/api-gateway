package v2alpha1

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
	v1beta1Status "github.com/kyma-project/api-gateway/internal/processing/status"
)

func (r Reconciliation) GetStatusBase(statusCode string) status.ReconciliationStatus {
	return v1beta1Status.ReconciliationV2alpha1Status{
		ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
			State: gatewayv2alpha1.State(statusCode),
		},
	}
}
