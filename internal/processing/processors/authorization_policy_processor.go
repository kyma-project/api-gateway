package processors

import (
	"context"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/go-logr/logr"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/hashbasedstate"
	"github.com/kyma-project/api-gateway/internal/subresources/authorizationpolicy"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthorizationPolicyProcessor is the generic processor that handles the Istio JwtAuthorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	ApiRule    *gatewayv1beta1.APIRule
	Creator    AuthorizationPolicyCreator
	Log        *logr.Logger
	Repository authorizationpolicy.Repository
}

// AuthorizationPolicyCreator provides the creation of AuthorizationPolicies using the configuration in the given APIRule.
type AuthorizationPolicyCreator interface {
	Create(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (hashbasedstate.Desired, error)
}

func (r AuthorizationPolicyProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired, err := r.getDesiredState(ctx, client, r.ApiRule)
	if err != nil {
		return nil, err
	}
	actual, err := r.getActualState(ctx, r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return changes, nil
}

func (r AuthorizationPolicyProcessor) getDesiredState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (hashbasedstate.Desired, error) {
	hashDummy, err := r.Creator.Create(ctx, client, api)
	if err != nil {
		return hashDummy, err
	}
	return hashDummy, nil
}

func (r AuthorizationPolicyProcessor) getActualState(ctx context.Context, api *gatewayv1beta1.APIRule) (hashbasedstate.Actual, error) {
	state := hashbasedstate.NewActual()

	apList, err := r.Repository.GetAll(ctx, api)
	if err != nil {
		return state, err
	}

	for _, ap := range apList {
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
