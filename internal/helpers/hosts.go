package helpers

import (
	"regexp"
)

const (
	fqdnMaxLength = 255
)

var (
	regexFqdn      = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9]{2,63}$`)
	regexShortName = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
)

func IsFqdnHostName(host string) bool {
	return len(host) <= fqdnMaxLength && regexFqdn.MatchString(host)
}

func IsShortHostName(host string) bool {
	return regexShortName.MatchString(host)
}
