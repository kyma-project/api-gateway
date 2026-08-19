package helpers

import (
	"encoding/json"
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

	return false, fmt.Sprintf("Status code %d on url %s is not between %d and %d", response.StatusCode, response.Request.URL, s.LowerStatusBound, s.UpperStatusBound)
}

// BodyHasHeaderValuePredicate checks that go-httpbin's /headers response
// contains each expected key-value pair. Header matching is case-insensitive.
type BodyHasHeaderValuePredicate struct {
	Expected [][2]string
}

func (b *BodyHasHeaderValuePredicate) Assert(response http.Response) (bool, string) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Sprintf("Failed to read response body from url %s", response.Request.URL)
	}

	var parsed struct {
		Headers map[string][]string `json:"headers"`
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return false, fmt.Sprintf("Failed to parse JSON body from url %s: %v", response.Request.URL, err)
	}

	var missing []string
	for _, kv := range b.Expected {
		key, wantVal := kv[0], kv[1]
		if !headersContain(parsed.Headers, key, wantVal) {
			missing = append(missing, fmt.Sprintf("%s=%s", key, wantVal))
		}
	}

	if len(missing) > 0 {
		return false, fmt.Sprintf("Response from %s missing headers: %s", response.Request.URL, strings.Join(missing, ", "))
	}
	return true, ""
}

func headersContain(headers map[string][]string, key, wantVal string) bool {
	for hk, vals := range headers {
		if strings.EqualFold(hk, key) {
			for _, v := range vals {
				if strings.EqualFold(v, wantVal) {
					return true
				}
			}
		}
	}
	return false
}

// BodyHasCookieValuePredicate checks that go-httpbin's /cookies response
// contains each expected key-value pair. The endpoint returns a flat
// JSON map of cookie name to value.
type BodyHasCookieValuePredicate struct {
	Expected [][2]string
}

func (b *BodyHasCookieValuePredicate) Assert(response http.Response) (bool, string) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Sprintf("Failed to read response body from url %s", response.Request.URL)
	}

	var cookies map[string]string
	if err := json.Unmarshal(bodyBytes, &cookies); err != nil {
		return false, fmt.Sprintf("Failed to parse JSON body from url %s: %v", response.Request.URL, err)
	}

	var missing []string
	for _, kv := range b.Expected {
		key, wantVal := kv[0], kv[1]
		if got, ok := cookies[key]; !ok || got != wantVal {
			missing = append(missing, fmt.Sprintf("%s=%s", key, wantVal))
		}
	}

	if len(missing) > 0 {
		return false, fmt.Sprintf("Response from %s missing cookies: %s", response.Request.URL, strings.Join(missing, ", "))
	}
	return true, ""
}

type BodyContainsPredicate struct {
	Expected []string
}

// Assert asserts that the response body contains the expected string
func (s *BodyContainsPredicate) Assert(response http.Response) (bool, string) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Sprintf("Failed to read response body from url %s", response.Request.URL)
	}

	bodyString := string(bodyBytes)

	var notContained []string
	for _, e := range s.Expected {
		if !strings.Contains(strings.ToLower(bodyString), strings.ToLower(e)) {
			notContained = append(notContained, e)
		}
	}

	if len(notContained) == 0 {
		return true, ""
	} else {
		return false, fmt.Sprintf("Body got from url %s didn't contain '%s'", response.Request.URL, strings.Join(notContained, "', '"))
	}

}
