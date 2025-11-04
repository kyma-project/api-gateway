package migration

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/subresources/accessrule"
)

type accessRuleDeletionProcessor struct {
	config     processing.ReconciliationConfig
	apiRule    *gatewayv1beta1.APIRule
	repository accessrule.Repository
}

func (a accessRuleDeletionProcessor) EvaluateReconciliation(ctx context.Context, _ client.Client) ([]*processing.ObjectChange, error) {
	ownedRules, err := a.repository.GetAll(ctx, a.apiRule)
	if err != nil {
		return nil, err
	}
	var changes []*processing.ObjectChange
	for _, rule := range ownedRules {
		changes = append(changes, processing.NewObjectDeleteAction(rule))
	}

	return changes, nil
}

// NewAccessRuleDeletionProcessor returns a new instance of the AccessRuleDeletionProcessor.
func NewAccessRuleDeletionProcessor(config processing.ReconciliationConfig, api *gatewayv1beta1.APIRule, k8sClient client.Client) processing.ReconciliationProcessor {
	return accessRuleDeletionProcessor{
		apiRule:    api,
		config:     config,
		repository: accessrule.NewRepository(k8sClient),
	}
}
