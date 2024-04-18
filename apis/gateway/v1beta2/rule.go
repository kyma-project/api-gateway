package v1beta2

import "encoding/json"

func (r *Rule) GetCookieMutator() (CookieMutatorConfig, error) {
	var mutatorConfig CookieMutatorConfig

	if r.Mutators == nil {
		return mutatorConfig, nil
	}

	for _, mutator := range r.Mutators {
		if mutator.Handler == CookieMutator {

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
		if mutator.Handler == HeaderMutator {
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

func (r *Rule) ContainsAccessStrategyJwt() bool {
	return r.Jwt != nil
}

func (r *Rule) ContainsNoAuth() bool {
	return r.NoAuth != nil
}
