package v1beta1

import (
	"fmt"
	"slices"

	gatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"github.com/kyma-project/api-gateway/internal/validation"
)

// CheckForExclusiveAccessStrategy checks if there is an access strategy that is not allowed in combination with other access strategies.
func CheckForExclusiveAccessStrategy(accessStrategies []*gatewayv1beta1.Authenticator, exclusiveAccessStrategy string, attributePath string) []validation.Failure {
	if len(accessStrategies) <= 1 {
		return nil
	}

	handlerIndex := slices.IndexFunc(accessStrategies, func(a *gatewayv1beta1.Authenticator) bool {
		return a.Name == exclusiveAccessStrategy
	})

	if handlerIndex > -1 {
		path := fmt.Sprintf("%s[%d]%s", attributePath+".accessStrategies", handlerIndex, ".handler")
		return []validation.Failure{
			{AttributePath: path, Message: fmt.Sprintf("%s access strategy is not allowed in combination with other access strategies", exclusiveAccessStrategy)},
		}
	}

	return nil
}

// CheckForSecureAndUnsecureAccessStrategies checks if there are secure and unsecure access strategies used at the same time.
func CheckForSecureAndUnsecureAccessStrategies(accessStrategies []*gatewayv1beta1.Authenticator, attributePath string) []validation.Failure {
	var containsSecureAccessStrategy, containsUnsecureAccessStrategy bool

	for _, r := range accessStrategies {
		switch r.Name {
		case gatewayv1beta1.AccessStrategyOauth2ClientCredentials,
			gatewayv1beta1.AccessStrategyOauth2Introspection,
			gatewayv1beta1.AccessStrategyJwt,
			gatewayv1beta1.AccessStrategyCookieSession:
			containsSecureAccessStrategy = true
		case gatewayv1beta1.AccessStrategyNoop,
			gatewayv1beta1.AccessStrategyNoAuth,
			gatewayv1beta1.AccessStrategyAllow,
			gatewayv1beta1.AccessStrategyUnauthorized,
			gatewayv1beta1.AccessStrategyAnonymous:
			containsUnsecureAccessStrategy = true
		}
	}

	if containsSecureAccessStrategy && containsUnsecureAccessStrategy {
		return []validation.Failure{{AttributePath: attributePath, Message: "Secure access strategies cannot be used in combination with unsecure access strategies"}}
	}

	return nil
}
