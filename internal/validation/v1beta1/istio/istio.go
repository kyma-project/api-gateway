package istio

import (
	"context"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"github.com/kyma-project/api-gateway/internal/validation/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
