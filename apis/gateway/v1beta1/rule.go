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
		if mutator.Handler.Name == CookieMutator {

			err := json.Unmarshal(mutator.Handler.Config.Raw, &mutatorConfig)

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
		if mutator.Handler.Name == HeaderMutator {
			err := json.Unmarshal(mutator.Handler.Config.Raw, &mutatorConfig)

			if err != nil {
				return mutatorConfig, err
			}

			return mutatorConfig, err
		}
	}

	return mutatorConfig, nil
}
