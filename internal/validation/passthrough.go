package validation

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

type passthrough struct{}

func (p *passthrough) Validate(config *runtime.RawExtension) error {
	if configNotEmpty(config) {
		return fmt.Errorf("passthrough mode requires empty configuration")
	}
	return nil
}
