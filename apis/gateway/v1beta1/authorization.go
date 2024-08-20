package v1beta1

func (a *JwtAuthorization) HasRequiredScopes() bool {
	return len(a.RequiredScopes) > 0
}
