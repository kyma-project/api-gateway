package v2alpha1

func (a *JwtAuthorization) HasRequiredScopes() bool {
	return a.RequiredScopes != nil && len(a.RequiredScopes) > 0
}
