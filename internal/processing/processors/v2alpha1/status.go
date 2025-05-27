package v2alpha1

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
)

func (r Reconciliation) GetStatusBase(state string) status.ReconciliationStatus {
	return status.ReconciliationV2alpha1Status{
		APIRuleStatus: &gatewayv2alpha1.APIRuleStatus{
			State: gatewayv2alpha1.State(state),
		},
	}
}
