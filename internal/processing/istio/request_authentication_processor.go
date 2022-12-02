package istio

import (
	"context"
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// RequestAuthenticationProcessor is the generic processor that handles the Istio Request Authentications in the reconciliation of API Rule.
type RequestAuthenticationProcessor struct {
	Creator RequestAuthenticationCreator
}

// RequestAuthenticationCreator provides the creation of Request Authentications using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type RequestAuthenticationCreator interface {
	Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.RequestAuthentication
}

// NewRequestAuthenticationProcessor returns a RequestAuthenticationProcessor with the desired state handling specific for the Istio handler.
func NewRequestAuthenticationProcessor(config processing.ReconciliationConfig) RequestAuthenticationProcessor {
	return RequestAuthenticationProcessor{
		Creator: requestAuthenticationCreator{
			additionalLabels: config.AdditionalLabels,
		},
	}
}

type requestAuthenticationCreator struct {
	additionalLabels map[string]string
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r requestAuthenticationCreator) Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.RequestAuthentication {
	requestAuthentications := make(map[string]*securityv1beta1.RequestAuthentication)
	for _, rule := range api.Spec.Rules {
		if processing.IsSecured(rule) {
			ra := generateRequestAuthentication(api, rule, r.additionalLabels)
			requestAuthentications[getRequestAuthenticationKey(ra)] = ra
		}
	}
	return requestAuthentications
}

func generateRequestAuthentication(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string) *securityv1beta1.RequestAuthentication {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := processing.GenerateOwnerRef(api)

	raBuilder := builders.RequestAuthenticationBuilder().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.RequestAuthenticationSpecBuilder().From(generateRequestAuthenticationSpec(api, rule))).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(processing.OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		raBuilder.Label(k, v)
	}

	return raBuilder.Get()
}

func generateRequestAuthenticationSpec(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *v1beta1.RequestAuthentication {

	var serviceName string
	if rule.Service != nil {
		serviceName = *rule.Service.Name
	} else {
		serviceName = *api.Spec.Service.Name
	}

	requestAuthenticationSpec := builders.RequestAuthenticationSpecBuilder().
		Selector(builders.SelectorBuilder().MatchLabels("app", serviceName)).
		JwtRules(builders.JwtRuleBuilder().From(rule.AccessStrategies))

	return requestAuthenticationSpec.Get()
}

func (r RequestAuthenticationProcessor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client, apiRule *gatewayv1beta1.APIRule) ([]*processing.ObjectChange, error) {
	desired := r.getDesiredState(apiRule)
	actual, err := r.getActualState(ctx, client, apiRule)
	if err != nil {
		return make([]*processing.ObjectChange, 0), err
	}

	changes := r.getObjectChanges(desired, actual)

	return changes, nil
}

func (r RequestAuthenticationProcessor) getDesiredState(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.RequestAuthentication {
	return r.Creator.Create(api)
}

func (r RequestAuthenticationProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	labels := processing.GetOwnerLabels(api)

	var raList securityv1beta1.RequestAuthenticationList
	if err := client.List(ctx, &raList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	requestAuthentications := make(map[string]*securityv1beta1.RequestAuthentication)

	for i := range raList.Items {
		obj := raList.Items[i]
		requestAuthentications[getRequestAuthenticationKey(obj)] = obj
	}

	return requestAuthentications, nil
}

func (r RequestAuthenticationProcessor) getObjectChanges(desiredRas map[string]*securityv1beta1.RequestAuthentication, actualRas map[string]*securityv1beta1.RequestAuthentication) []*processing.ObjectChange {
	raChanges := make(map[string]*processing.ObjectChange)

	for path, rule := range desiredRas {

		if actualRas[path] != nil {
			actualRas[path].Spec = rule.Spec
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

func getRequestAuthenticationKey(ra *securityv1beta1.RequestAuthentication) string {
	key := ""
	for _, k := range ra.Spec.JwtRules {
		key += fmt.Sprintf("%s:%s", k.Issuer, k.JwksUri)
	}
	return key
}
