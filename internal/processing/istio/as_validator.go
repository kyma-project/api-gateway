package istio

import (
	"fmt"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"golang.org/x/exp/slices"
)

type asValidator struct{}

func (o *asValidator) Validate(attributePath string, accessStrategies []*gatewayv1beta1.Authenticator) []validation.Failure {
	var problems []validation.Failure

	if len(accessStrategies) > 1 {
		allowIndex := slices.IndexFunc(accessStrategies, func(a *gatewayv1beta1.Authenticator) bool { return a.Handler.Name == "allow" })
		jwtIndex := slices.IndexFunc(accessStrategies, func(a *gatewayv1beta1.Authenticator) bool { return a.Handler.Name == "jwt" })
		if allowIndex > -1 {
			attrPath := fmt.Sprintf("%s[%d]%s", attributePath+".accessStrategies", allowIndex, ".handler")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: "allow access strategy is not allowed in combination with other access strategies"})
		}
		if jwtIndex > -1 {
			attrPath := fmt.Sprintf("%s[%d]%s", attributePath+".accessStrategies", jwtIndex, ".handler")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: "jwt access strategy is not allowed in combination with other access strategies"})
		}
	}

	return problems
}
