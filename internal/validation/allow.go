package validation

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type allow struct{}

func (a *allow) Validate(gate *gatewayv2alpha1.Gate) error {
	if len(gate.Spec.Rules) != 1 {
		return fmt.Errorf("supplied config should contain exactly one path")
	}
	if hasDuplicates(gate.Spec.Rules) {
		return fmt.Errorf("supplied config is invalid: multiple definitions of the same path detected")
	}
	return nil
}
