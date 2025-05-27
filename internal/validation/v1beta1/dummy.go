package v1beta1

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

// dummy is a handler validator that does nothing.
type dummyHandlerValidator struct{}

func (dummy *dummyHandlerValidator) Validate(_ string, _ *gatewayv1beta1.Handler) []validation.Failure {
	return nil
}

type dummyAccessStrategiesValidator struct{}

func (dummy *dummyAccessStrategiesValidator) Validate(_ string, _ []*gatewayv1beta1.Authenticator) []validation.Failure {
	return nil
}
