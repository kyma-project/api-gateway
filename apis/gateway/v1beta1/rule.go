package v1beta1

import "encoding/json"

func (r *Rule) GetJwtIstioAuthorizations() []*JwtAuthorization {
	// For Istio JWT we can safely assume that there is only one access strategy
	accessStrategy := r.AccessStrategies[0]

	authorizations := &JwtConfig{
		Authorizations: []*JwtAuthorization{},
	}
	if accessStrategy.Config != nil {
		_ = json.Unmarshal(accessStrategy.Config.Raw, authorizations)
	}

	return authorizations.Authorizations
}

func (r *Rule) GetCookieMutator() (CookieMutatorConfig, error) {
	var mutatorConfig CookieMutatorConfig

	if r.Mutators == nil {
		return mutatorConfig, nil
	}

	for _, mutator := range r.Mutators {
		if mutator.Name == CookieMutator {
			err := json.Unmarshal(mutator.Config.Raw, &mutatorConfig)

			if err != nil {
				return mutatorConfig, err
			}

			return mutatorConfig, nil
		}
	}

	return mutatorConfig, nil
}

func (r *Rule) GetHeaderMutator() (HeaderMutatorConfig, error) {
	var mutatorConfig HeaderMutatorConfig

	if r.Mutators == nil {
		return mutatorConfig, nil
	}

	for _, mutator := range r.Mutators {
		if mutator.Name == HeaderMutator {
			err := json.Unmarshal(mutator.Config.Raw, &mutatorConfig)

			if err != nil {
				return mutatorConfig, err
			}

			return mutatorConfig, err
		}
	}

	return mutatorConfig, nil
}

func ConvertHttpMethodsToStrings(methods []HttpMethod) []string {
	strings := make([]string, len(methods))
	for i, method := range methods {
		strings[i] = string(method)
	}

	return strings
}

const (
	AccessStrategyAllow                   string = "allow"
	AccessStrategyNoAuth                  string = "no_auth"
	AccessStrategyJwt                     string = "jwt"
	AccessStrategyNoop                    string = "noop"
	AccessStrategyUnauthorized            string = "unauthorized"
	AccessStrategyAnonymous               string = "anonymous"
	AccessStrategyCookieSession           string = "cookie_session"
	AccessStrategyOauth2ClientCredentials string = "oauth2_client_credentials"
	AccessStrategyOauth2Introspection     string = "oauth2_introspection"
)

func (r *Rule) ContainsAccessStrategy(strategy string) bool {
	for _, s := range r.AccessStrategies {
		if s.Name == strategy {
			return true
		}
	}
	return false
}
