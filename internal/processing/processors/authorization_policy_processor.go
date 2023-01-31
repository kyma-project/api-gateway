package processors

import (
	"context"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const AuthorizationPolicyAppSelectorLabel = "app"

// AuthorizationPolicyProcessor is the generic processor that handles the Istio JwtAuthorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	Creator AuthorizationPolicyCreator
}

// AuthorizationPolicyCreator provides the creation of AuthorizationPolicies using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type AuthorizationPolicyCreator interface {
	Create(api *gatewayv1beta1.APIRule) []*securityv1beta1.AuthorizationPolicy
}

func (r AuthorizationPolicyProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return changes, nil
}

func (r AuthorizationPolicyProcessor) getDesiredState(api *gatewayv1beta1.APIRule) []*securityv1beta1.AuthorizationPolicy {
	return r.Creator.Create(api)
}

func (r AuthorizationPolicyProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) ([]*securityv1beta1.AuthorizationPolicy, error) {
	labels := processing.GetOwnerLabels(api)

	var apList securityv1beta1.AuthorizationPolicyList
	if err := client.List(ctx, &apList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	authorizationPolicies := append([]*securityv1beta1.AuthorizationPolicy{}, apList.Items...)

	return authorizationPolicies, nil
}

func (r AuthorizationPolicyProcessor) getObjectChanges(desiredAps []*securityv1beta1.AuthorizationPolicy, actualAps []*securityv1beta1.AuthorizationPolicy) []*processing.ObjectChange {
	var apObjectActionsToApply []*processing.ObjectChange

	for _, ap := range desiredAps {
		objectAction := processing.NewObjectCreateAction(ap)
		apObjectActionsToApply = append(apObjectActionsToApply, objectAction)
	}

	for _, ap := range actualAps {
		objectAction := processing.NewObjectDeleteAction(ap)
		apObjectActionsToApply = append(apObjectActionsToApply, objectAction)
	}

	return apObjectActionsToApply
}
