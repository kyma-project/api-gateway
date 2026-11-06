package v2alpha1

import (
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing/status"
)

func (r Reconciliation) GetStatusBase(state string) status.ReconciliationStatus {
	generation := int64(0)
	if r.apiRuleV2alpha1 != nil {
		generation = r.apiRuleV2alpha1.Generation
	}

	return status.ReconciliationV2alpha1Status{
		ApiRuleStatus: &gatewayv2alpha1.APIRuleStatus{
			State: gatewayv2alpha1.State(state),
		},
		Generation:       generation,
		ManagedResources: managedResources(r.apiRuleV2alpha1),
	}
}

// managedResources reports which sub-resources are managed for the APIRule: VS and AP
// always, RequestAuthentication only when a rule requests JWT or ext-auth restrictions
// (mirrors the predicate in requestauthentication/creator.go).
func managedResources(apiRule *gatewayv2alpha1.APIRule) map[status.ResourceSelector]bool {
	managed := map[status.ResourceSelector]bool{
		status.OnVirtualService:      true,
		status.OnAuthorizationPolicy: true,
	}
	if apiRule == nil {
		return managed
	}
	for _, rule := range apiRule.Spec.Rules {
		if rule.Jwt != nil || rule.ExtAuth != nil && rule.ExtAuth.Restrictions != nil {
			managed[status.OnRequestAuthentication] = true
			break
		}
	}
	return managed
}
