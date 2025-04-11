package istio

import (
	"context"
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/helpers"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Newv1beta1RequestAuthenticationProcessor returns a RequestAuthenticationProcessor with the desired state handling specific for the Istio handler.
func Newv1beta1RequestAuthenticationProcessor(config processing.ReconciliationConfig, apiRule *gatewayv1beta1.APIRule) processors.RequestAuthenticationProcessor {
	return processors.RequestAuthenticationProcessor{
		ApiRule: apiRule,
		Creator: requestAuthenticationCreator{},
	}
}

type requestAuthenticationCreator struct{}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r requestAuthenticationCreator) Create(ctx context.Context, client client.Client, api *gatewayv1beta1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	requestAuthentications := make(map[string]*securityv1beta1.RequestAuthentication)
	for _, rule := range api.Spec.Rules {
		if processing.IsJwtSecured(rule) {
			ra, err := generateRequestAuthentication(ctx, client, api, rule)
			if err != nil {
				return requestAuthentications, err
			}
			requestAuthentications[processors.GetRequestAuthenticationKey(ra)] = ra
		}
	}
	return requestAuthentications, nil
}

func generateRequestAuthentication(ctx context.Context, client client.Client, api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) (*securityv1beta1.RequestAuthentication, error) {
	namePrefix := fmt.Sprintf("%s-", api.Name)
	namespace := helpers.FindServiceNamespace(api, &rule)

	spec, err := generateRequestAuthenticationSpec(ctx, client, api, rule)
	if err != nil {
		return nil, err
	}

	raBuilder := builders.NewRequestAuthenticationBuilder().
		WithGenerateName(namePrefix).
		WithNamespace(namespace).
		WithSpec(builders.NewRequestAuthenticationSpecBuilder().WithFrom(spec).Get()).
		WithLabel(processing.OwnerLabel, fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	return raBuilder.Get(), nil
}

func generateRequestAuthenticationSpec(ctx context.Context, client client.Client, api *gatewayv1beta1.APIRule, rule gatewayv1beta1.Rule) (*v1beta1.RequestAuthentication, error) {
	var service *gatewayv1beta1.Service
	if rule.Service != nil {
		service = rule.Service
	} else {
		service = api.Spec.Service
	}

	labelSelector, err := helpers.GetLabelSelectorFromService(ctx, client, service, api, &rule)
	if err != nil {
		return nil, err
	}

	requestAuthenticationSpec := builders.NewRequestAuthenticationSpecBuilder().
		WithSelector(labelSelector).
		WithJwtRules(*builders.NewJwtRuleBuilder().From(rule.AccessStrategies).Get())

	return requestAuthenticationSpec.Get(), nil
}
