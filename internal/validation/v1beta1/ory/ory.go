package ory

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"github.com/kyma-project/api-gateway/internal/validation/v1beta1"
)

func NewAPIRuleValidator(api *gatewayv1beta1.APIRule, defaultDomainName string) validation.APIRuleValidator {
	return &v1beta1.APIRuleValidator{
		APIRule: api,

		HandlerValidator:          &handlerValidator{},
		AccessStrategiesValidator: &accessStrategyValidator{},
		DefaultDomainName:         defaultDomainName,
	}
}
