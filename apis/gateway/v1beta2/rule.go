package v1beta2

func (r *Rule) ContainsAccessStrategyJwt() bool {
	return r.Jwt != nil
}

func (r *Rule) ContainsNoAuth() bool {
	return r.NoAuth != nil
}
