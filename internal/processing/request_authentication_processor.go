package processing

import (
	"context"
	"fmt"

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
	Create(api *gatewayv1beta1.APIRule) map[string]*istiosecurityv1beta1.RequestAuthentication
}

func (r RequestAuthenticationProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule) ([]*ObjectChange, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return changes, nil
}

func (r RequestAuthenticationProcessor) getDesiredState(api *gatewayv1beta1.APIRule) map[string]*istiosecurityv1beta1.RequestAuthentication {
	return r.Creator.Create(api)
}

func (r RequestAuthenticationProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (map[string]*istiosecurityv1beta1.RequestAuthentication, error) {
	labels := GetOwnerLabels(api)

	var raList istiosecurityv1beta1.RequestAuthenticationList
	if err := client.List(ctx, &raList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	requestAuthentications := make(map[string]*istiosecurityv1beta1.RequestAuthentication)

	for i := range raList.Items {
		obj := raList.Items[i]
		requestAuthentications[GetRequestAuthenticationKey(obj)] = obj
	}

	return requestAuthentications, nil
}

func (r RequestAuthenticationProcessor) getObjectChanges(desiredRas map[string]*istiosecurityv1beta1.RequestAuthentication, actualRas map[string]*istiosecurityv1beta1.RequestAuthentication) []*ObjectChange {
	raChanges := make(map[string]*ObjectChange)

	for path, rule := range desiredRas {

		if actualRas[path] != nil {
			actualRas[path].Spec = rule.Spec
			raChanges[path] = NewObjectUpdateAction(actualRas[path])
		} else {
			raChanges[path] = NewObjectCreateAction(rule)
		}

	}

	for path, rule := range actualRas {
		if desiredRas[path] == nil {
			raChanges[path] = NewObjectDeleteAction(rule)
		}
	}

	raChangesToApply := make([]*ObjectChange, 0, len(raChanges))

	for _, applyCommand := range raChanges {
		raChangesToApply = append(raChangesToApply, applyCommand)
	}

	return raChangesToApply
}

func GetRequestAuthenticationKey(ra *istiosecurityv1beta1.RequestAuthentication) string {
	key := ""
	for _, k := range ra.Spec.JwtRules {
		key += fmt.Sprintf("%s:%s", k.Issuer, k.JwksUri)
	}
	return key
}
