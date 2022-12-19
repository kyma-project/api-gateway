package validation

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

// dummy is an handler validator that does nothing
type dummyHandlerValidator struct{}

func (dummy *dummyHandlerValidator) Validate(attrPath string, handler *gatewayv1beta1.Handler) []Failure {
	return nil
}

type dummyAccessStrategiesValidator struct{}

func (dummy *dummyAccessStrategiesValidator) Validate(attrPath string, accessStrategies []*gatewayv1beta1.Authenticator) []Failure {
	return nil
}
