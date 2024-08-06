package validation

import (
	"bytes"
	"errors"
	"net/url"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
)

func IsInvalidURI(toTest string) (bool, error) {
	if len(toTest) == 0 {
		return true, errors.New("value is empty")
	}
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return true, err
	}
	return false, nil
}

// ValidateDomainName ?
func ValidateDomainName(domain string) bool {
	RegExp := regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-_]*\.)*[a-zA-Z0-9]*[a-zA-Z0-9-_]*[[a-zA-Z0-9]+$`)
	return RegExp.MatchString(domain)
}

// ValidateSubdomainName ?
func ValidateSubdomainName(subdomain string) bool {
	RegExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return RegExp.MatchString(subdomain)
}

// ValidateServiceName ?
func ValidateServiceName(service string) bool {
	regExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?\.[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return regExp.MatchString(service)
}

func validateGatewayName(gateway string) bool {
	regExp := regexp.MustCompile(`^[0-9a-z-_]+(\/[0-9a-z-_]+|(\.[0-9a-z-_]+)*)$`)
	return regExp.MatchString(gateway)
}

// configNotEmpty Verify if the config object is not empty
func configEmpty(config *runtime.RawExtension) bool {

	return config == nil ||
		len(config.Raw) == 0 ||
		bytes.Equal(config.Raw, []byte("null")) ||
		bytes.Equal(config.Raw, []byte("{}"))
}

// ConfigNotEmpty Verify if the config object is not empty
func ConfigNotEmpty(config *runtime.RawExtension) bool {
	return !configEmpty(config)
}
