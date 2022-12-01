package processing

import (
	"context"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// RequestAuthenticationProcessor is the generic processor that handles the Istio Request Authentications in the reconciliation of API Rule.
type RequestAuthenticationProcessor struct {
	Creator RequestAuthenticationCreator
}

// RequestAuthenticationCreator provides the creation of Request Authentications using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type RequestAuthenticationCreator interface {
	Create(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *istiosecurityv1beta1.RequestAuthentication
}

func (r RequestAuthenticationProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) ([]*ObjectChange, error) {
	desired := r.getDesiredState(apiRule, rule)
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return []*ObjectChange{changes}, nil
}

func (r RequestAuthenticationProcessor) getDesiredState(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *istiosecurityv1beta1.RequestAuthentication {
	return r.Creator.Create(api, rule)
}

func (r RequestAuthenticationProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (*istiosecurityv1beta1.RequestAuthentication, error) {
	labels := GetOwnerLabels(api)

	var raList istiosecurityv1beta1.RequestAuthenticationList
	if err := client.List(ctx, &raList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	if len(raList.Items) >= 1 {
		return raList.Items[0], nil
	} else {
		return nil, nil
	}
}

func (r RequestAuthenticationProcessor) getObjectChanges(desiredRa *istiosecurityv1beta1.RequestAuthentication, actualRa *istiosecurityv1beta1.RequestAuthentication) *ObjectChange {
	if actualRa != nil {
		actualRa.Spec = *desiredRa.Spec.DeepCopy()
		return NewObjectUpdateAction(actualRa)
	} else {
		return NewObjectCreateAction(desiredRa)
	}
}
