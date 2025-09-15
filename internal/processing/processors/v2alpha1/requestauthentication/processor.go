package requestauthentication

import (
	"context"
	"fmt"
	"strings"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/api-gateway/internal/processing"
)

// NewProcessor returns a processor with the desired state handling specific for the Istio handler.
func NewProcessor(apiRule *gatewayv2alpha1.APIRule) Processor {
	return Processor{
		ApiRule: apiRule,
		Creator: requestAuthenticationCreator{},
	}
}

// Processor is the generic processor that handles the Istio Request Authentications in the reconciliation of API Rule.
type Processor struct {
	ApiRule *gatewayv2alpha1.APIRule
	Creator requestAuthenticationCreator
}

func (r Processor) EvaluateReconciliation(ctx context.Context, client ctrlclient.Client) ([]*processing.ObjectChange, error) {
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

func (r Processor) getDesiredState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	return r.Creator.Create(ctx, client, api)
}

func (r Processor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv2alpha1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	labels := processing.GetOwnerLabelsV2alpha1(api)

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

func (r Processor) getObjectChanges(desiredRas map[string]*securityv1beta1.RequestAuthentication, actualRas map[string]*securityv1beta1.RequestAuthentication) []*processing.ObjectChange {
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

func getSelectorsKey(labels map[string]string) string {
	var selector = ""
	for kay, value := range labels {
		selector += fmt.Sprintf("%s=%s,", kay, value)
	}
	return strings.TrimRight(selector, ",")
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
		getSelectorsKey(ra.Spec.Selector.MatchLabels),
		jwtRulesKey,
		// If the namespace changed, the resource should be recreated
		namespace,
	)
}
