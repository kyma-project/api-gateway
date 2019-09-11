package validation

import (
	"encoding/json"
	"fmt"
	"net/url"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/pkg/errors"
)

type jwt struct{}

func (j *jwt) Validate(gate *gatewayv2alpha1.Gate) error {
	var template gatewayv2alpha1.JWTModeConfig

	if len(gate.Spec.Paths) == 0 {
		return fmt.Errorf("path is required")
	}

	if !configNotEmpty(gate.Spec.Auth.Config) {
		return fmt.Errorf("supplied config cannot be empty")
	}
	err := json.Unmarshal(gate.Spec.Auth.Config.Raw, &template)
	if err != nil {
		return errors.WithStack(err)
	}
	if !j.isValidURL(template.Issuer) {
		return fmt.Errorf("issuer field is empty or not a valid url")
	}
	return nil
}

func (j *jwt) isValidURL(toTest string) bool {
	if len(toTest) == 0 {
		return false
	}
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}
