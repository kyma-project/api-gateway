package authorizationpolicy

import (
	"context"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/subresources/authorizationpolicy"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"

	"github.com/go-logr/logr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
)

// NewProcessor returns a Processor with the desired state handling for AuthorizationPolicy.
func NewProcessor(log *logr.Logger, rule *gatewayv2alpha1.APIRule, gateway *networkingv1beta1.Gateway, client ctrlclient.Client) Processor {
	return Processor{
		apiRule:    rule,
		creator:    creator{gateway: gateway},
		Log:        log,
		repository: authorizationpolicy.NewRepository(client),
	}
}

// NewMigrationProcessor returns a Processor with the desired state handling for AuthorizationPolicy when in the migration process from v1beta1 to v2alpha1.
func NewMigrationProcessor(log *logr.Logger, rule *gatewayv2alpha1.APIRule, oryPassthrough bool, gateway *networkingv1beta1.Gateway, client ctrlclient.Client) Processor {
	return Processor{
		apiRule: rule,
		creator: creator{
			oryPassthrough: oryPassthrough,
			gateway:        gateway,
		},
		Log:        log,
		repository: authorizationpolicy.NewRepository(client),
	}
}

// Processor handles the Istio AuthorizationPolicy in the reconciliation of API Rule.
type Processor struct {
	apiRule    *gatewayv2alpha1.APIRule
	creator    Creator
	Log        *logr.Logger
	repository authorizationpolicy.Repository
}

func (p Processor) EvaluateReconciliation(ctx context.Context, k8sClient ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired, err := p.getDesiredState(ctx, k8sClient, p.apiRule)
	if err != nil {
		return nil, err
	}
	actual, err := p.getActualState(ctx, k8sClient, p.apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := p.getObjectChanges(desired, actual)

	return changes, nil
}

func (p Processor) getDesiredState(ctx context.Context, k8sClient ctrlclient.Client, api *gatewayv2alpha1.APIRule) (hashbasedstate.Desired, error) {
	hashDummy, err := p.creator.Create(ctx, k8sClient, api)
	if err != nil {
		return hashDummy, err
	}
	return hashDummy, nil
}

func (p Processor) getActualState(ctx context.Context, _ ctrlclient.Client, api *gatewayv2alpha1.APIRule) (hashbasedstate.Actual, error) {
	state := hashbasedstate.NewActual()

	authorizationPolicies, err := p.repository.GetAll(ctx, api)
	if err != nil {
		return state, err
	}
	for _, ap := range authorizationPolicies {
		h := hashbasedstate.NewAuthorizationPolicy(ap)
		state.Add(&h)
	}

	return state, nil
}

func (p Processor) getObjectChanges(desired hashbasedstate.Desired, actual hashbasedstate.Actual) []*processing.ObjectChange {
	var apObjectActionsToApply []*processing.ObjectChange

	changes := hashbasedstate.GetChanges(desired, actual)
	p.Log.Info("Authorization policy changes that will be applied", "changes", changes)

	for _, ap := range changes.Create {
		apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectCreateAction(ap))
	}

	for _, ap := range changes.Update {
		apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectUpdateAction(ap))
	}

	for _, ap := range changes.Delete {
		apObjectActionsToApply = append(apObjectActionsToApply, processing.NewObjectDeleteAction(ap))
	}

	return apObjectActionsToApply
}
