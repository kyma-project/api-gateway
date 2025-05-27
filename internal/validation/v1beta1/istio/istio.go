package istio

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"github.com/kyma-project/api-gateway/internal/validation/v1beta1"
)

func NewAPIRuleValidator(ctx context.Context, client client.Client, api *gatewayv1beta1.APIRule, defaultDomainName string) validation.ApiRuleValidator {
	return &v1beta1.APIRuleValidator{
		ApiRule: api,

		HandlerValidator:          &handlerValidator{},
		AccessStrategiesValidator: &accessStrategyValidator{},
		MutatorsValidator:         &mutatorsValidator{},
		InjectionValidator:        validation.NewInjectionValidator(ctx, client),
		RulesValidator:            &RulesValidator{},
		DefaultDomainName:         defaultDomainName,
	}
}
