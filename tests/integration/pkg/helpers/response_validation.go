package helpers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HttpResponseAsserter interface {
	// Assert asserts that the response is valid and returns true if it is. It also returns a message with details about the failure.
	Assert(response http.Response) (bool, string)
}

// StatusPredicate is a struct representing desired endpoint call response status code, that is between LowerStatusBound and UpperStatusBound
type StatusPredicate struct {
	LowerStatusBound int
	UpperStatusBound int
}

func (s *StatusPredicate) Assert(response http.Response) (bool, string) {
	if response.StatusCode >= s.LowerStatusBound && response.StatusCode <= s.UpperStatusBound {
		return true, ""
	}

	return false, fmt.Sprintf("Status code %d is not between %d and %d", response.StatusCode, s.LowerStatusBound, s.UpperStatusBound)
}

// BodyContainsPredicate is a struct representing desired HTTP response body containing expected strings
type BodyContainsPredicate struct {
	Expected []string
}

// Assert asserts that the response body contains the expected string
func (s *BodyContainsPredicate) Assert(response http.Response) (bool, string) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, "Failed to read response body"
	}

	bodyString := string(bodyBytes)

	var notContained []string
	for _, e := range s.Expected {
		if !strings.Contains(bodyString, e) {
			notContained = append(notContained, e)
		}
	}

	if len(notContained) == 0 {
		return true, ""
	} else {
		return false, fmt.Sprintf("Body didn't contain '%s'", strings.Join(notContained, "', '"))
	}

}
