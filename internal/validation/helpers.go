package validation

import (
	"fmt"
	"net/url"
	"regexp"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
)

func hasDuplicates(rules []gatewayv1alpha1.Rule) bool {
	duplicates := map[string]bool{}

	if len(rules) > 1 {
		for _, rule := range rules {
			if len(rule.Methods) > 0 {
				for _, method := range rule.Methods {
					tmp := fmt.Sprintf("%s:%s", rule.Path, method)
					if duplicates[tmp] {
						return true
					}
					duplicates[tmp] = true
				}
			} else {
				if duplicates[rule.Path] {
					return true
				}
				duplicates[rule.Path] = true
			}
		}
	}

	return false
}

func isValidURL(toTest string) bool {
	if len(toTest) == 0 {
		return false
	}
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}

//ValidateDomainName ?
func ValidateDomainName(domain string) bool {
	RegExp := regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-_]*\.)*[a-zA-Z0-9]*[a-zA-Z0-9-_]*[[a-zA-Z0-9]+$`)
	return RegExp.MatchString(domain)
}

//ValidateServiceName ?
func ValidateServiceName(service string) bool {
	regExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?\.[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return regExp.MatchString(service)
}
