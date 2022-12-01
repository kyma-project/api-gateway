package istio

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"github.com/kyma-incubator/api-gateway/internal/processing"
	"istio.io/api/security/v1beta1"
	istiosecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

// NewRequestAuthenticationProcessor returns a RequestAuthenticationProcessor with the desired state handling specific for the Istio handler.
func NewRequestAuthenticationProcessor(config processing.ReconciliationConfig) processing.RequestAuthenticationProcessor {
	return processing.RequestAuthenticationProcessor{
		Creator: requestAuthenticationCreator{
			oathkeeperSvc:     config.OathkeeperSvc,
			oathkeeperSvcPort: config.OathkeeperSvcPort,
			corsConfig:        config.CorsConfig,
			additionalLabels:  config.AdditionalLabels,
			defaultDomainName: config.DefaultDomainName,
		},
	}
}

type requestAuthenticationCreator struct {
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	corsConfig        *processing.CorsConfig
	defaultDomainName string
	additionalLabels  map[string]string
}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r requestAuthenticationCreator) Create(api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) *istiosecurityv1beta1.RequestAuthentication {
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

	for k, v := range r.additionalLabels {
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
