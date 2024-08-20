package shared

func (a *JwtAuthorization) HasRequiredScopes() bool {
	return a.RequiredScopes != nil && len(a.RequiredScopes) > 0
}
