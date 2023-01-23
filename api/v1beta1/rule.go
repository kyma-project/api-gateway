package v1beta1

import "encoding/json"

func (r *Rule) GetAuthorizations() []*Authorization {
	// We assume only one accessStrategy
	accessStrategy := r.AccessStrategies[0]

	authorizations := &Authorizations{
		Authorizations: []*Authorization{},
	}
	if accessStrategy.Config != nil {
		_ = json.Unmarshal(accessStrategy.Config.Raw, authorizations)
	}

	return authorizations.Authorizations
}
