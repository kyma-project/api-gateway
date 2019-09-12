package validation

import (
	"net/url"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

func hasDuplicates(paths []gatewayv2alpha1.Rule) bool {
	encountered := map[string]bool{}
	// Create a map of all unique elements.
	for v := range paths {
		encountered[paths[v].Path] = true
	}
	return len(encountered) != len(paths)
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
