package validation

import (
	"encoding/json"
	"fmt"
	"net/url"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	jwtModes = []string{gatewayv2alpha1.JWTAll, gatewayv2alpha1.JWTInclude, gatewayv2alpha1.JWTExclude}
)

type jwt struct{}

func (j *jwt) Validate(config *runtime.RawExtension) error {
	var template gatewayv2alpha1.JWTModeConfig

	if !configNotEmpty(config) {
		return fmt.Errorf("supplied config cannot be empty")
	}
	err := json.Unmarshal(config.Raw, &template)
	if err != nil {
		return errors.WithStack(err)
	}
	if !j.isValidURL(template.Issuer) {
		return fmt.Errorf("issuer field is empty or not a valid url")
	}
	if !j.isValidMode(template.Mode.Name) {
		return fmt.Errorf("supplied mode is invalid: %v, valid modes are: ALL, INCLUDE, EXCLUDE", template.Mode.Name)
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

func (j *jwt) isValidMode(mode string) bool {
	for _, b := range jwtModes {
		if b == mode {
			return true
		}
	}
	return false
}
