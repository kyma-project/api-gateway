package v1beta1

import "strings"

const (
	CookieMutator = "cookie"
	HeaderMutator = "header"
)

type CookieMutatorConfig struct {
	Cookies map[string]string `json:"cookies"`
}

func (c CookieMutatorConfig) ToString() string {
	var cookies []string
	for name, value := range c.Cookies {
		cookies = append(cookies, name+"="+value)
	}

	return strings.Join(cookies, "; ")
}

type HeaderMutatorConfig struct {
	Headers map[string]string `json:"headers"`
}
