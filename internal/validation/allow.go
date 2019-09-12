package validation

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type allow struct{}

func (a *allow) Validate(gate *gatewayv2alpha1.Gate) error {
	if len(gate.Spec.Paths) != 1 {
		return fmt.Errorf("supplied config should contain exactly one path")
	}
	if hasDuplicates(gate.Spec.Paths) {
		return fmt.Errorf("supplied config is invalid: multiple definitions of the same path detected")
	}
	if len(gate.Spec.Paths[0].Scopes) > 0 {
		return fmt.Errorf("allow mode does not support scopes")
	}
	return nil
}
