package v2alpha1

func (a *JwtAuthorization) HasRequiredScopes() bool {
	return len(a.RequiredScopes) > 0
}
