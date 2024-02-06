package istio

import (
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"github.com/kyma-project/api-gateway/internal/validation"
	"golang.org/x/exp/slices"
)

type asValidator struct{}

func (o *asValidator) Validate(attributePath string, accessStrategies []*gatewayv1beta1.Authenticator) []validation.Failure {
	var problems []validation.Failure

	if len(accessStrategies) > 1 {
		allowIndex := slices.IndexFunc(accessStrategies, func(a *gatewayv1beta1.Authenticator) bool {
			return a.Handler.Name == gatewayv1beta1.AccessStrategyAllow
		})
		allowMethodsIndex := slices.IndexFunc(accessStrategies, func(a *gatewayv1beta1.Authenticator) bool {
			return a.Handler.Name == gatewayv1beta1.AccessStrategyAllowMethods
		})
		jwtIndex := slices.IndexFunc(accessStrategies, func(a *gatewayv1beta1.Authenticator) bool { return a.Handler.Name == gatewayv1beta1.AccessStrategyJwt })
		if allowIndex > -1 {
			attrPath := fmt.Sprintf("%s[%d]%s", attributePath+".accessStrategies", allowIndex, ".handler")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: "allow access strategy is not allowed in combination with other access strategies"})
		}
		if allowMethodsIndex > -1 {
			attrPath := fmt.Sprintf("%s[%d]%s", attributePath+".accessStrategies", allowMethodsIndex, ".handler")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: fmt.Sprintf("%s access strategy is not allowed in combination with other access strategies", gatewayv1beta1.AccessStrategyAllowMethods)})
		}
		if jwtIndex > -1 {
			attrPath := fmt.Sprintf("%s[%d]%s", attributePath+".accessStrategies", jwtIndex, ".handler")
			problems = append(problems, validation.Failure{AttributePath: attrPath, Message: "jwt access strategy is not allowed in combination with other access strategies"})
		}
	}

	return problems
}
