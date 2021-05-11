package validation

import gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

//dummy is an accessStrategy validator that does nothing
type dummyAccStrValidator struct{}

func (dummy *dummyAccStrValidator) Validate(attrPath string, handler *gatewayv1alpha1.Handler) []Failure {
	return nil
}
