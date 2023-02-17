package helpers

import (
	"io"
	"net/http"
	"strings"
)

type HttpResponseAsserter interface {
	Assert(response *http.Response) bool
}

// StatusPredicate is a struct representing desired endpoint call response status code, that is between LowerStatusBound and UpperStatusBound
type StatusPredicate struct {
	LowerStatusBound int
	UpperStatusBound int
}

func (s *StatusPredicate) Assert(response *http.Response) bool {
	return response.StatusCode >= s.LowerStatusBound && response.StatusCode <= s.UpperStatusBound
}

// BodyContainsPredicate is a struct representing desired HTTP response body containing an expected string
type BodyContainsPredicate struct {
	Expected string
}

// Assert asserts that the response body contains the expected string
func (s *BodyContainsPredicate) Assert(response *http.Response) bool {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false
	}
	bodyString := string(bodyBytes)
	return strings.Contains(bodyString, s.Expected)
}
