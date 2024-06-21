package v2alpha1

func (r *Rule) ContainsAccessStrategyJwt() bool {
	return r.Jwt != nil
}

func (r *Rule) ContainsNoAuth() bool {
	return r.NoAuth != nil
}
