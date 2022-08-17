package validation

import gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"

//dummy is an accessStrategy validator that does nothing
type dummyAccStrValidator struct{}

func (dummy *dummyAccStrValidator) Validate(attrPath string, handler *gatewayv1beta1.Handler) []Failure {
	return nil
}
