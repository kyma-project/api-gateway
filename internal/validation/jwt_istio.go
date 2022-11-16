package validation

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

type IstioJwtValidator struct{}

func (i *IstioJwtValidator) Validate(_ string, _ *gatewayv1beta1.Handler) []Failure {
	var problems []Failure

	return problems
}
