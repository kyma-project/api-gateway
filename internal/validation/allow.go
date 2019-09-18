package validation

import (
	"fmt"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
)

type allow struct{}

func (a *allow) Validate(api *gatewayv1alpha1.APIRule) error {
	if len(api.Spec.Rules) != 1 {
		return fmt.Errorf("supplied config should contain exactly one path")
	}
	if hasDuplicates(api.Spec.Rules) {
		return fmt.Errorf("supplied config is invalid: multiple definitions of the same path detected")
	}
	return nil
}
