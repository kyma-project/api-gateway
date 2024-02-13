package istio

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/kyma-project/api-gateway/internal/validation"
)

type asValidator struct{}

var exclusiveAccessStrategies = []string{gatewayv1beta1.AccessStrategyAllow, gatewayv1beta1.AccessStrategyNoAuth, gatewayv1beta1.AccessStrategyJwt}

func (o *asValidator) Validate(attributePath string, accessStrategies []*gatewayv1beta1.Authenticator) []validation.Failure {
	var problems []validation.Failure

	for _, strategy := range exclusiveAccessStrategies {
		validationProblems := validation.CheckForExclusiveAccessStrategy(accessStrategies, strategy, attributePath)
		problems = append(problems, validationProblems...)
	}

	return problems
}
