package processors

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	"github.com/kyma-project/api-gateway/internal/processing"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const RequestAuthenticationAppSelectorLabel = "app"

// RequestAuthenticationProcessor is the generic processor that handles the Istio Request Authentications in the reconciliation of API Rule.
type RequestAuthenticationProcessor struct {
	ApiRule *gatewayv2alpha1.APIRule
	Creator RequestAuthenticationCreator
}

// RequestAuthenticationCreator provides the creation of RequestAuthentications using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type RequestAuthenticationCreator interface {
	Create(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error)
}

func (r RequestAuthenticationProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
	desired, err := r.getDesiredState(ctx, client, r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}
	actual, err := r.getActualState(ctx, client, r.ApiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return changes, nil
}

func (r RequestAuthenticationProcessor) getDesiredState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	return r.Creator.Create(ctx, client, api)
}

func (r RequestAuthenticationProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	labels := processing.GetOwnerLabels(api)

	var raList securityv1beta1.RequestAuthenticationList
	if err := client.List(ctx, &raList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	requestAuthentications := make(map[string]*securityv1beta1.RequestAuthentication)

	for i := range raList.Items {
		obj := raList.Items[i]
		requestAuthentications[GetRequestAuthenticationKey(obj)] = obj
	}

	return requestAuthentications, nil
}

func (r RequestAuthenticationProcessor) getObjectChanges(desiredRas map[string]*securityv1beta1.RequestAuthentication, actualRas map[string]*securityv1beta1.RequestAuthentication) []*processing.ObjectChange {
	raChanges := make(map[string]*processing.ObjectChange)

	for path, rule := range desiredRas {

		if actualRas[path] != nil {
			actualRas[path].Spec = *rule.Spec.DeepCopy()
			raChanges[path] = processing.NewObjectUpdateAction(actualRas[path])
		} else {
			raChanges[path] = processing.NewObjectCreateAction(rule)
		}

	}

	for path, rule := range actualRas {
		if desiredRas[path] == nil {
			raChanges[path] = processing.NewObjectDeleteAction(rule)
		}
	}

	raChangesToApply := make([]*processing.ObjectChange, 0, len(raChanges))

	for _, applyCommand := range raChanges {
		raChangesToApply = append(raChangesToApply, applyCommand)
	}

	return raChangesToApply
}

func GetRequestAuthenticationKey(ra *securityv1beta1.RequestAuthentication) string {
	jwtRulesKey := ""

	for _, k := range ra.Spec.JwtRules {
		jwtRulesKey += fmt.Sprintf("%s:%s", k.Issuer, k.JwksUri)
	}

	namespace := ra.Namespace
	if namespace == "" {
		namespace = "default"
	}

	return fmt.Sprintf("%s:%s:%s",
		ra.Spec.Selector.MatchLabels[RequestAuthenticationAppSelectorLabel],
		jwtRulesKey,
		// If the namespace changed, the resource should be recreated
		namespace,
	)
}
