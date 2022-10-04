package validation

import (
	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/helpers"
)

// dummy is an accessStrategy validator that does nothing
type dummyAccStrValidator struct{}

func (dummy *dummyAccStrValidator) Validate(attrPath string, handler *gatewayv1beta1.Handler, config *helpers.Config) []Failure {
	return nil
}
