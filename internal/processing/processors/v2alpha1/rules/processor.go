package rules

import (
	"context"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/subresources/accessrule"

	"github.com/go-logr/logr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
)

// NewDeletionProcessor returns a Processor that only removes orphaned Ory Rules
func NewDeletionProcessor(log *logr.Logger, rule *gatewayv2alpha1.APIRule, client ctrlclient.Client) Processor {
	return Processor{
		apiRule:    rule,
		creator:    NoOpCreator{},
		Log:        log,
		repository: accessrule.NewRepository(client),
	}
}

type NoOpCreator struct{}

func (n NoOpCreator) Create(context.Context, ctrlclient.Client, *gatewayv2alpha1.APIRule) (hashbasedstate.Desired, error) {
	return hashbasedstate.NewDesired(), nil
}

type Processor struct {
	apiRule    *gatewayv2alpha1.APIRule
	creator    NoOpCreator
	Log        *logr.Logger
	repository accessrule.Repository
}

func (p Processor) EvaluateReconciliation(ctx context.Context, _ ctrlclient.Client) ([]*processing.ObjectChange, error) {
	rules, err := p.repository.GetAll(ctx, p.apiRule)
	if err != nil {
		return nil, err
	}
	changes := make([]*processing.ObjectChange, len(rules))
	for i, rule := range rules {
		changes[i] = processing.NewObjectDeleteAction(rule)
	}

	if len(changes) != 0 {
		p.Log.Info("Deleting orphaned Ory Rules", "number", len(changes))
	}
	return changes, nil
}
