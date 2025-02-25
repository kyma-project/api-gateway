package v2

func (r *Rule) ContainsAccessStrategyJwt() bool {
	return r.Jwt != nil
}

func (r *Rule) ContainsNoAuth() bool {
	return r.NoAuth != nil
}

func ConvertHttpMethodsToStrings(methods []HttpMethod) []string {
	strings := make([]string, len(methods))
	for i, method := range methods {
		strings[i] = string(method)
	}

	return strings
}

// AppliesToAllPaths returns true if the rule applies to all paths of the service.
func (r *Rule) AppliesToAllPaths() bool {
	return r.Path == "/*"
}
