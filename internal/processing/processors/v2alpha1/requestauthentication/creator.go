package requestauthentication

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"

	"github.com/kyma-project/api-gateway/internal/builders"
	"github.com/kyma-project/api-gateway/internal/processing"
	"github.com/kyma-project/api-gateway/internal/processing/processors"
	"istio.io/api/security/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type requestAuthenticationCreator struct{}

// Create returns the Virtual Service using the configuration of the APIRule.
func (r requestAuthenticationCreator) Create(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule) (map[string]*securityv1beta1.RequestAuthentication, error) {
	requestAuthentications := make(map[string]*securityv1beta1.RequestAuthentication)
	for _, rule := range api.Spec.Rules {
		if rule.Jwt != nil || rule.ExtAuth != nil && rule.ExtAuth.Restrictions != nil {
			ra, err := generateRequestAuthentication(ctx, client, api, rule)
			if err != nil {
				return requestAuthentications, err
			}
			requestAuthentications[processors.GetRequestAuthenticationKey(ra)] = ra
		}
	}
	return requestAuthentications, nil
}

func generateRequestAuthentication(ctx context.Context, client client.Client, apiRule *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (*securityv1beta1.RequestAuthentication, error) {
	namePrefix := fmt.Sprintf("%s-", apiRule.Name)
	namespace, err := gatewayv2alpha1.FindServiceNamespace(apiRule, rule)
	if err != nil {
		return nil, fmt.Errorf("finding service namespace: %w", err)
	}

	spec, err := generateRequestAuthenticationSpec(ctx, client, apiRule, rule)
	if err != nil {
		return nil, err
	}

	raBuilder := builders.NewRequestAuthenticationBuilder().
		WithGenerateName(namePrefix).
		WithNamespace(namespace).
		WithSpec(builders.NewRequestAuthenticationSpecBuilder().WithFrom(spec).Get()).
		WithLabel(processing.OwnerLabel, fmt.Sprintf("%s.%s", apiRule.Name, apiRule.Namespace)).
		WithLabel(processing.ModuleLabelKey, processing.ApiGatewayLabelValue).
		WithLabel(processing.K8sManagedByLabelKey, processing.ApiGatewayLabelValue).
		WithLabel(processing.K8sComponentLabelKey, processing.ApiGatewayLabelValue).
		WithLabel(processing.K8sPartOfLabelKey, processing.ApiGatewayLabelValue)

	return raBuilder.Get(), nil
}

func generateRequestAuthenticationSpec(ctx context.Context, client client.Client, api *gatewayv2alpha1.APIRule, rule gatewayv2alpha1.Rule) (*v1beta1.RequestAuthentication, error) {

	s, err := gatewayv2alpha1.GetSelectorFromService(ctx, client, api, rule)
	if err != nil {
		return nil, err
	}

	requestAuthenticationSpec := builders.NewRequestAuthenticationSpecBuilder().
		WithSelector(s.Selector)

	if rule.ExtAuth != nil && rule.ExtAuth.Restrictions != nil {
		requestAuthenticationSpec.WithJwtRules(*builders.NewJwtRuleBuilder().FromV2Alpha1(rule.ExtAuth.Restrictions).Get())
	} else {
		requestAuthenticationSpec.WithJwtRules(*builders.NewJwtRuleBuilder().FromV2Alpha1(rule.Jwt).Get())
	}

	return requestAuthenticationSpec.Get(), nil
}
