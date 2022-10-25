package processing

import (
	"fmt"
	"istio.io/api/security/v1beta1"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

func modifyRequestAuthentication(existing, required *securityv1beta1.RequestAuthentication) {
	existing.Spec = *required.Spec.DeepCopy()
}

func generateRequestAuthentication(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule, additionalLabels map[string]string) *securityv1beta1.RequestAuthentication {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := generateOwnerRef(api)

	arBuilder := builders.RequestAuthenticationBuilder().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.RequestAuthenticationSpecBuilder().From(generateRequestAuthenticationSpec(rule))).
		Label(OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range additionalLabels {
		arBuilder.Label(k, v)
	}

	return arBuilder.Get()
}

func generateRequestAuthenticationSpec(rule gatewayv1beta1.Rule) *v1beta1.RequestAuthentication {
	requestAuthenticationSpec := builders.RequestAuthenticationSpecBuilder().
		Selector(builders.SelectorBuilder().MatchLabels("app", *rule.Service.Name)).
		JwtRules(builders.JwtRuleBuilder())

	return requestAuthenticationSpec.Get()
}
