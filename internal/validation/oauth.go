package validation

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type oauth struct{}

func (o *oauth) Validate(gate *gatewayv2alpha1.Gate) error {
	if len(gate.Spec.Paths) != 1 {
		return fmt.Errorf("supplied config should contain exactly one path")
	}
	if o.hasDuplicates(gate.Spec.Paths) {
		return fmt.Errorf("supplied config is invalid: multiple definitions of the same path detected")
	}
	return nil
}

func (o *oauth) hasDuplicates(paths []gatewayv2alpha1.Path) bool {
	encountered := map[string]bool{}
	// Create a map of all unique elements.
	for v := range paths {
		encountered[paths[v].Path] = true
	}
	return len(encountered) != len(paths)
}
