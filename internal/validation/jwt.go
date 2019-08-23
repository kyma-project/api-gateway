package validation

import "k8s.io/apimachinery/pkg/runtime"

type jwt struct{}

func (j *jwt) Validate(config *runtime.RawExtension) error {
	return nil
}
