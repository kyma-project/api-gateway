package istio

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// NewAuthorizationPolicyProcessor returns a AuthorizationPolicyProcessor with the desired state handling specific for the Istio handler.
func NewAuthorizationPolicyProcessor(config processing.ReconciliationConfig) processing.AuthorizationPolicyProcessor {
	return processing.AuthorizationPolicyProcessor{
		Creator: authorizationPolicyCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type authorizationPolicyCreator struct {
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	defaultDomainName string
	additionalLabels  map[string]string
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r authorizationPolicyCreator) Create(api *gatewayv1beta1.APIRule) map[string]*securityv1beta1.AuthorizationPolicy {
	namePrefix := fmt.Sprintf("%s-", api.ObjectMeta.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := processing.GenerateOwnerRef(api)

	arBuilder := builders.AuthorizationPolicyBuilder().
		GenerateName(namePrefix).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AuthorizationPolicySpecBuilder().From(generateAuthorizationPolicySpec(api, rule))).
		Label(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)).
		Label(processing.OwnerLabelv1alpha1, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	for k, v := range r.additionalLabels {
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
