package migration

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type accessRuleDeletionProcessor struct {
	config  processing.ReconciliationConfig
	apiRule *gatewayv2alpha1.APIRule
}

func (a accessRuleDeletionProcessor) EvaluateReconciliation(ctx context.Context, k8sClient client.Client) ([]*processing.ObjectChange, error) {
	var ownedRules rulev1alpha1.RuleList
	labels := processing.GetOwnerLabels(a.apiRule)
	if err := k8sClient.List(ctx, &ownedRules, client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	var changes []*processing.ObjectChange
	for _, rule := range ownedRules.Items {
		changes = append(changes, processing.NewObjectDeleteAction(&rule))
	}

	return changes, nil
}

// NewAccessRuleDeletionProcessor returns a new instance of the AccessRuleDeletionProcessor.
func NewAccessRuleDeletionProcessor(config processing.ReconciliationConfig, api *gatewayv2alpha1.APIRule) processing.ReconciliationProcessor {
	return accessRuleDeletionProcessor{
		apiRule: api,
		config:  config,
	}
}
