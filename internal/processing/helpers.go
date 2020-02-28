package processing

import (
	"fmt"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyAccessRule(existing, required *rulev1alpha1.Rule) {
	existing.Spec = required.Spec
}

func generateAccessRule(api *gatewayv1alpha1.APIRule, rule gatewayv1alpha1.Rule, accessStrategies []*rulev1alpha1.Authenticator, additionalLabels map[string]string) *rulev1alpha1.Rule {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := generateOwnerRef(api)

	arBuilder := builders.AccessRule().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AccessRuleSpec().From(generateAccessRuleSpec(api, rule, accessStrategies))).
		Label(OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		arBuilder.Label(k, v)
	}

	return arBuilder.Get()
}

func generateAccessRuleSpec(api *gatewayv1alpha1.APIRule, rule gatewayv1alpha1.Rule, accessStrategies []*rulev1alpha1.Authenticator) *rulev1alpha1.RuleSpec {
	return builders.AccessRuleSpec().
		Upstream(builders.Upstream().
			URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)))).
		Match(builders.Match().
			URL(fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, rule.Path)).
			Methods(rule.Methods)).
		Authorizer(builders.Authorizer().Handler(builders.Handler().
			Name("allow"))).
		Authenticators(builders.Authenticators().From(accessStrategies)).
		Mutators(builders.Mutators().From(rule.Mutators)).Get()
}

func isSecured(rule gatewayv1alpha1.Rule) bool {
	if len(rule.Mutators) > 0 {
		return true
	}
	for _, strat := range rule.AccessStrategies {
		if strat.Name != "allow" {
			return true
		}
	}
	return false
}

func generateOwnerRef(api *gatewayv1alpha1.APIRule) k8sMeta.OwnerReference {
	return *builders.OwnerReference().
		Name(api.ObjectMeta.Name).
		APIVersion(api.TypeMeta.APIVersion).
		Kind(api.TypeMeta.Kind).
		UID(api.ObjectMeta.UID).
		Controller(true).
		Get()
}
