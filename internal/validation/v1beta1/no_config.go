package v1beta1

import (
	"bytes"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

// noConfig is an accessStrategy validator that does not accept nested config.
type noConfigAccStrValidator struct{}

func (a *noConfigAccStrValidator) Validate(attrPath string, handler *gatewayv1beta1.Handler) []validation.Failure {
	var problems []validation.Failure

	if handler.Config != nil && len(handler.Config.Raw) > 0 && !bytes.Equal(handler.Config.Raw, []byte("null")) && !bytes.Equal(handler.Config.Raw, []byte("{}")) {
		problems = append(problems, validation.Failure{AttributePath: attrPath + ".config", Message: "strategy: " + handler.Name + " does not support configuration"})
	}

	return problems
}
