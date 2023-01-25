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
