package istio

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
)

type jwtValidator struct{}

func (i *jwtValidator) Validate(_ string, _ *gatewayv1beta1.Handler) []validation.Failure {
	var problems []validation.Failure

	return problems
}
