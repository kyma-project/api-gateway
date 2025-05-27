package processors

import (
	"context"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthorizationPolicyProcessor is the generic processor that handles the Istio JwtAuthorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	ApiRule *gatewayv2alpha1.APIRule
	Creator AuthorizationPolicyCreator
	Log     *logr.Logger
}

// AuthorizationPolicyCreator provides the creation of AuthorizationPolicies using the configuration in the given APIRule.
type AuthorizationPolicyCreator interface {
	Create(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (hashbasedstate.Desired, error)
}

func (r AuthorizationPolicyProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired, err := r.getDesiredState(ctx, client, r.ApiRule)
	if err != nil {
		return nil, err
	}
	actual, err := r.getActualState(ctx, client, r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return changes, nil
}

func (r AuthorizationPolicyProcessor) getDesiredState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (hashbasedstate.Desired, error) {
	hashDummy, err := r.Creator.Create(ctx, client, api)
	if err != nil {
		return hashDummy, err
	}
	return hashDummy, nil
}

func (r AuthorizationPolicyProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (hashbasedstate.Actual, error) {
	state := hashbasedstate.NewActual()

	labels := processing.GetOwnerLabels(api)

	var apList securityv1beta1.AuthorizationPolicyList
	if err := client.List(ctx, &apList, ctrlclient.MatchingLabels(labels)); err != nil {
		return state, err
	}

	for _, ap := range apList.Items {
		h := hashbasedstate.NewAuthorizationPolicy(ap)
		state.Add(&h)
	}

	return state, nil
}

func (r AuthorizationPolicyProcessor) getObjectChanges(desired hashbasedstate.Desired, actual hashbasedstate.Actual) []*processing.ObjectChange {
	var apObjectActionsToApply []*processing.ObjectChange

	changes := hashbasedstate.GetChanges(desired, actual)
	r.Log.Info("Authorization policy changes that will be applied", "changes", changes)

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
