package istio

import (
	"fmt"

	"istio.io/api/security/v1beta1"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

func modifyAuthorizationPolicy(existing, required *securityv1beta1.AuthorizationPolicy) {
	existing.Spec = *required.Spec.DeepCopy()
}

func generateAuthorizationPolicy(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string) *securityv1beta1.AuthorizationPolicy {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := GenerateOwnerRef(api)

	arBuilder := builders.AuthorizationPolicyBuilder().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AuthorizationPolicySpecBuilder().From(generateAuthorizationPolicySpec(api, rule))).
		Label(OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		arBuilder.Label(k, v)
	}

	return arBuilder.Get()
}

func generateAuthorizationPolicySpec(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *v1beta1.AuthorizationPolicy {
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
