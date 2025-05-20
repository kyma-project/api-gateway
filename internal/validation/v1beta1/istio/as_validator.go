package istio

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
	"github.com/kyma-project/api-gateway/internal/validation/v1beta1"
)

type accessStrategyValidator struct{}

var exclusiveAccessStrategies = []string{
	gatewayv1beta1.AccessStrategyAllow,
	gatewayv1beta1.AccessStrategyNoAuth,
	gatewayv1beta1.AccessStrategyJwt,
	gatewayv1beta1.AccessStrategyNoop,
}

func (o *accessStrategyValidator) Validate(attributePath string, accessStrategies []*gatewayv1beta1.Authenticator) []validation.Failure {
	var problems []validation.Failure

	for _, strategy := range exclusiveAccessStrategies {
		validationProblems := v1beta1.CheckForExclusiveAccessStrategy(accessStrategies, strategy, attributePath)
		problems = append(problems, validationProblems...)
	}

	return problems
}
