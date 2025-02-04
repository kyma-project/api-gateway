package v2

func (a *JwtAuthorization) HasRequiredScopes() bool {
	return len(a.RequiredScopes) > 0
}
