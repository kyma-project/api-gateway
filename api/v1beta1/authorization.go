package v1beta1

func (a *Authorization) HasRequiredScopes() bool {
	return a.RequiredScopes != nil && len(a.RequiredScopes) > 0
}
