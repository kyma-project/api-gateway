package helpers

import (
	"fmt"
	"strings"
)

func GetHostWithDomain(host, defaultDomainName string) string {
	if !HostIncludesDomain(host) {
		return GetHostWithDefaultDomain(host, defaultDomainName)
	}
	return host
}

func HostIncludesDomain(host string) bool {
	return strings.Contains(host, ".")
}

func GetHostWithDefaultDomain(host, defaultDomainName string) string {
	return fmt.Sprintf("%s.%s", host, defaultDomainName)
}

func GetHostLocalDomain(host string, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", host, namespace)
}
