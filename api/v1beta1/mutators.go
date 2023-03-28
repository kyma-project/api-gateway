package v1beta1

import (
  "strings"
  "fmt"
)

const (
	CookieMutator = "cookie"
	HeaderMutator = "header"
)

type CookieMutatorConfig struct {
	Cookies map[string]string `json:"cookies"`
}

func (c *CookieMutatorConfig) HasCookies() bool {
	return c.Cookies != nil && len(c.Cookies) > 0
}

func (c *CookieMutatorConfig) ToString() string {
	var cookies []string
	for name, value := range c.Cookies {
		cookies = append(cookies, fmt.Sprintf("%s=%s,name,value))
	}

	return strings.Join(cookies, "; ")
}

type HeaderMutatorConfig struct {
	Headers map[string]string `json:"headers"`
}

func (h *HeaderMutatorConfig) HasHeaders() bool {
	return h.Headers != nil && len(h.Headers) > 0
}
