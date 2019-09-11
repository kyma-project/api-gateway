package validation

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type passthrough struct{}

func (p *passthrough) Validate(gate *gatewayv2alpha1.Gate) error {
	if len(gate.Spec.Paths) != 0 {
		return fmt.Errorf("passthrough mode requires empty configuration")
	}
	return nil
}
