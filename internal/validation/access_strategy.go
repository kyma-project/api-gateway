package validation

import (
	"fmt"
	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"

	"golang.org/x/exp/slices"
)

func CheckForExclusiveAccessStrategy(accessStrategies []*gatewayv1beta1.Authenticator, exclusiveAccessStrategy string, attributePath string) []Failure {

	if len(accessStrategies) <= 1 {
		return nil
	}

	handlerIndex := slices.IndexFunc(accessStrategies, func(a *gatewayv1beta1.Authenticator) bool {
		return a.Handler.Name == exclusiveAccessStrategy
	})

	if handlerIndex > -1 {
		path := fmt.Sprintf("%s[%d]%s", attributePath+".accessStrategies", handlerIndex, ".handler")
		return []Failure{{AttributePath: path, Message: fmt.Sprintf("%s access strategy is not allowed in combination with other access strategies", exclusiveAccessStrategy)}}
	}

	return nil
}
