package processing

import (
	"fmt"

	"github.com/kyma-incubator/api-gateway/internal/helpers"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
)

func modifyAccessRule(existing, required *rulev1alpha1.Rule) {
	existing.Spec = required.Spec
}

func generateAccessRule(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, accessStrategies []*gatewayv1beta1.Authenticator, additionalLabels map[string]string, defaultDomainName string) *rulev1alpha1.Rule {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := generateOwnerRef(api)

	arBuilder := builders.AccessRule().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AccessRuleSpec().From(generateAccessRuleSpec(api, rule, accessStrategies, defaultDomainName))).
		Label(OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		arBuilder.Label(k, v)
	}

	return arBuilder.Get()
}

func generateAccessRuleSpec(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, accessStrategies []*gatewayv1beta1.Authenticator, defaultDomainName string) *rulev1alpha1.RuleSpec {
	accessRuleSpec := builders.AccessRuleSpec().
		Match(builders.Match().
			URL(fmt.Sprintf("<http|https>://%s<%s>", helpers.GetHostWithDomain(*api.Spec.Host, defaultDomainName), rule.Path)).
			Methods(rule.Methods)).
		Authorizer(builders.Authorizer().Handler(builders.Handler().
			Name("allow"))).
		Authenticators(builders.Authenticators().From(accessStrategies)).
		Mutators(builders.Mutators().From(rule.Mutators))

	// Use rule level service if it exists
	if rule.Service != nil {
		return accessRuleSpec.Upstream(builders.Upstream().
			URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *rule.Service.Name, api.ObjectMeta.Namespace, int(*rule.Service.Port)))).Get()
	}
	// Otherwise use service defined on APIRule spec level
	return accessRuleSpec.Upstream(builders.Upstream().
		URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)))).Get()
}
