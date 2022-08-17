package validation

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	gatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
)

func hasDuplicates(rules []gatewayv1beta1.Rule) bool {
	encountered := map[string]bool{}
	// Create a map of all unique elements.
	for v := range rules {
		encountered[rules[v].Path] = true
	}
	return len(encountered) != len(rules)
}

func isInvalidURL(toTest string) (bool, error) {
	if len(toTest) == 0 {
		return true, errors.New("value is empty")
	}
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return true, err
	}
	return false, nil
}

func isUnsecuredURL(toTest string) (bool, error) {
	if len(toTest) == 0 {
		return true, errors.New("value is empty")
	}
	if strings.HasPrefix(toTest, "http://") {
		return true, errors.New("value is unsecure")
	}
	return false, nil
}

//ValidateDomainName ?
func ValidateDomainName(domain string) bool {
	RegExp := regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-_]*\.)*[a-zA-Z0-9]*[a-zA-Z0-9-_]*[[a-zA-Z0-9]+$`)
	return RegExp.MatchString(domain)
}

//ValidateSubdomainName ?
func ValidateSubdomainName(subdomain string) bool {
	RegExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return RegExp.MatchString(subdomain)
}

//ValidateServiceName ?
func ValidateServiceName(service string) bool {
	regExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?\.[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return regExp.MatchString(service)
}

func validateGatewayName(gateway string) bool {
	regExp := regexp.MustCompile(`^[0-9a-z-_]+(\/[0-9a-z-_]+|(\.[0-9a-z-_]+)*)$`)
	return regExp.MatchString(gateway)
}
