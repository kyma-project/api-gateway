package v1beta2

func (a *JwtAuthorization) HasRequiredScopes() bool {
	return a.RequiredScopes != nil && len(a.RequiredScopes) > 0
}
