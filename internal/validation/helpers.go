package validation

import (
	"net/url"
	"regexp"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
)

func hasDuplicates(rules []gatewayv1alpha1.Rule) bool {
	encountered := map[string]bool{}
	// Create a map of all unique elements.
	for v := range rules {
		encountered[rules[v].Path] = true
	}
	return len(encountered) != len(rules)
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
