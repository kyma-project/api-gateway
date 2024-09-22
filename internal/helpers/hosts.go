package helpers

import (
	"regexp"
)

const (
	fqdnMaxLength = 255
)

var (
	regexFqdn      = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)
	regexShortName = regexp.MustCompile("^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$")
)

func IsHostFqdn(host string) bool {
	return len(host) <= fqdnMaxLength && regexFqdn.MatchString(host)
}

func IsHostShortName(host string) bool {
	return regexShortName.MatchString(host)
}
