package validation

import (
	"bytes"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

//noConfig is an accessStrategy validator that does not accept nested config
type noConfigAccStrValidator struct{}

func (a *noConfigAccStrValidator) Validate(attrPath string, handler *gatewayv1beta1.Handler) []Failure {
	var problems []Failure

	if handler.Config != nil && len(handler.Config.Raw) > 0 && !bytes.Equal(handler.Config.Raw, []byte("null")) && !bytes.Equal(handler.Config.Raw, []byte("{}")) {
		problems = append(problems, Failure{AttributePath: attrPath + ".config", Message: "strategy: " + handler.Name + " does not support configuration"})
	}

	return problems
}
