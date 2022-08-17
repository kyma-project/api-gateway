package validation

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

type oauth struct{}

func (o *oauth) Validate(api *gatewayv1beta1.APIRule) error {
	if len(api.Spec.Rules) != 1 {
		return fmt.Errorf("supplied config should contain exactly one path")
	}
	if hasDuplicates(api.Spec.Rules) {
		return fmt.Errorf("supplied config is invalid: multiple definitions of the same path detected")
	}
	return nil
}
