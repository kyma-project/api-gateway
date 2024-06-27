package migration

import (
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

type handlerValidator struct{}

func (o handlerValidator) Validate(_ string, _ *gatewayv1beta1.Handler) []validation.Failure {
	var failures []validation.Failure

	// TODO
	return failures
}
