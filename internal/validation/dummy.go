package validation

import "github.com/ory/oathkeeper-maester/api/v1alpha1"

//dummy is an accessStrategy validator that does nothing
type dummyAccStrValidator struct{}

func (dummy *dummyAccStrValidator) Validate(attrPath string, handler *v1alpha1.Handler) []Failure {
	return nil
}
