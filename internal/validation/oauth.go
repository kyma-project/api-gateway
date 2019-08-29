package validation

import (
	"encoding/json"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type oauth struct{}

func (o *oauth) Validate(config *runtime.RawExtension) error {
	var template gatewayv2alpha1.OauthModeConfig

	if !configNotEmpty(config) {
		return fmt.Errorf("supplied config cannot be empty")
	}

	//Check if the supplied data is castable to OauthModeConfig
	err := json.Unmarshal(config.Raw, &template)
	if err != nil {
		return errors.WithStack(err)
	}
	// If not, the result is an empty template object.
	// Check if template is empty
	if len(template.Paths) != 1 {
		return fmt.Errorf("supplied config should contain exactly one path")
	}
	if o.hasDuplicates(template.Paths) {
		return fmt.Errorf("supplied config is invalid: multiple definitions of the same path detected")
	}
	return nil
}

func (o *oauth) hasDuplicates(elements []gatewayv2alpha1.Option) bool {
	encountered := map[string]bool{}
	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v].Path] = true
	}
	return len(encountered) != len(elements)
}
