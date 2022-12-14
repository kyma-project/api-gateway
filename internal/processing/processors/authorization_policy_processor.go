package processors

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

// AuthorizationPolicyProcessor is the generic processor that handles the Istio Authorization Policies in the reconciliation of API Rule.
type AuthorizationPolicyProcessor struct {
	Creator AuthorizationPolicyCreator
}

// AuthorizationPolicyCreator provides the creation of AuthorizationPolicies using the configuration in the given APIRule.
// The key of the map is expected to be unique and comparable with the
type AuthorizationPolicyCreator interface {
	Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.AuthorizationPolicy
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

func (r AuthorizationPolicyProcessor) getDesiredState(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.AuthorizationPolicy {
	return r.Creator.Create(api)
}

func (r AuthorizationPolicyProcessor) getActualState(ctx context.Context, client ctrlclient.Client, api *gatewayv1beta1.APIRule) (map[string]*securityv1beta1.AuthorizationPolicy, error) {
	labels := processing.GetOwnerLabels(api)

	var apList securityv1beta1.AuthorizationPolicyList
	if err := client.List(ctx, &apList, ctrlclient.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	authorizationPolicies := make(map[string]*securityv1beta1.AuthorizationPolicy)
	pathDuplicates := HasPathDuplicates(api.Spec.Rules)
	for i := range apList.Items {
		obj := apList.Items[i]
		authorizationPolicies[GetAuthorizationPolicyKey(pathDuplicates, obj)] = obj
	}

	return authorizationPolicies, nil
}

func (r AuthorizationPolicyProcessor) getObjectChanges(desiredAps map[string]*securityv1beta1.AuthorizationPolicy, actualAps map[string]*securityv1beta1.AuthorizationPolicy) []*processing.ObjectChange {
	apChanges := make(map[string]*processing.ObjectChange)

	for path, rule := range desiredAps {

		if actualAps[path] != nil {
			actualAps[path].Spec = *rule.Spec.DeepCopy()
			apChanges[path] = processing.NewObjectUpdateAction(actualAps[path])
		} else {
			apChanges[path] = processing.NewObjectCreateAction(rule)
		}

	}

	for path, rule := range actualAps {
		if desiredAps[path] == nil {
			apChanges[path] = processing.NewObjectDeleteAction(rule)
		}
	}

	apChangesToApply := make([]*processing.ObjectChange, 0, len(apChanges))

	for _, applyCommand := range apChanges {
		apChangesToApply = append(apChangesToApply, applyCommand)
	}

	return apChangesToApply
}

func GenerateAuthorizationPolicy(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string) *securityv1beta1.AuthorizationPolicy {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := processing.GenerateOwnerRef(api)

	apBuilder := builders.AuthorizationPolicyBuilder().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AuthorizationPolicySpecBuilder().From(GenerateAuthorizationPolicySpec(api, rule))).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(processing.OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		apBuilder.Label(k, v)
	}

	return apBuilder.Get()
}

func GenerateAuthorizationPolicySpec(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *v1beta1.AuthorizationPolicy {
	var serviceName string
	if rule.Service != nil {
		serviceName = *rule.Service.Name
	} else {
		serviceName = *api.Spec.Service.Name
	}

	authorizationPolicySpec := builders.AuthorizationPolicySpecBuilder().
		Selector(builders.SelectorBuilder().MatchLabels("app", serviceName)).
		Rule(builders.RuleBuilder().
			RuleFrom(builders.RuleFromBuilder().Source()).
			RuleTo(builders.RuleToBuilder().
				Operation(builders.OperationBuilder().Methods(rule.Methods).Path(rule.Path))))

	return authorizationPolicySpec.Get()
}

func GetAuthorizationPolicyKey(hasPathDuplicates bool, ap *securityv1beta1.AuthorizationPolicy) string {
	key := ""
	if ap.Spec.Rules != nil && len(ap.Spec.Rules) > 0 && ap.Spec.Rules[0].To != nil && len(ap.Spec.Rules[0].To) > 0 {
		if hasPathDuplicates {
			key = fmt.Sprintf("%s:%s",
				processing.SliceToString(ap.Spec.Rules[0].To[0].Operation.Paths),
				processing.SliceToString(ap.Spec.Rules[0].To[0].Operation.Methods))
		} else {
			key = processing.SliceToString(ap.Spec.Rules[0].To[0].Operation.Paths)
		}
	}

	return key
}
