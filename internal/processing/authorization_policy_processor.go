package processing

import (
	"context"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AuthorizationPolicyProcessor is the generic processor that handles the Istio Authorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	Creator AuthorizationPolicyCreator
}

// AuthorizationPolicyCreator provides the creation of Authorization Policies using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type AuthorizationPolicyCreator interface {
	Create(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *securityv1beta1.AuthorizationPolicy
}

func (r AuthorizationPolicyProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) ([]*ObjectChange, error) {
	desired := r.getDesiredState(apiRule, rule)
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return []*ObjectChange{changes}, nil
}

func (r AuthorizationPolicyProcessor) getDesiredState(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *securityv1beta1.AuthorizationPolicy {
	return r.Creator.Create(api, rule)
}

func (r AuthorizationPolicyProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (*securityv1beta1.AuthorizationPolicy, error) {
	labels := GetOwnerLabels(api)

	var arList securityv1beta1.AuthorizationPolicyList
	if err := client.List(ctx, &arList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(arList.Items) >= 1 {
		return arList.Items[0], nil
	} else {
		return nil, nil
	}
}

func (r AuthorizationPolicyProcessor) getObjectChanges(desiredRa *securityv1beta1.AuthorizationPolicy, actualRa *securityv1beta1.AuthorizationPolicy) *ObjectChange {
	if actualRa != nil {
		actualRa.Spec = *desiredRa.Spec.DeepCopy()
		return NewObjectUpdateAction(actualRa)
	} else {
		return NewObjectCreateAction(desiredRa)
	}
}
