package istiojwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

func initMutatorsOverwrite(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-mutators-overwrite.yaml", "istio-jwt-mutators-overwrite")

	ctx.Step(`^JwtMutatorsOverwrite: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(
		`^JwtMutatorsOverwrite: There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)"$`,
		scenario.thereIsAnEndpointWithHeaderMutator,
	)
	ctx.Step(
		`^JwtMutatorsOverwrite: There is an endpoint on path "([^"]*)" with a cookie mutator setting "([^"]*)" cookie to "([^"]*)"$`,
		scenario.thereIsAnEndpointWithCookieMutator,
	)
	ctx.Step(`^JwtMutatorsOverwrite: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^JwtMutatorsOverwrite: Calling the "([^"]*)" endpoint with a request having cookie header with value "([^"]*)" should return cookie header with value "([^"]*)"$`,
		scenario.shouldOverwriteCookieValue,
	)
	ctx.Step(
		`^JwtMutatorsOverwrite: Calling the "([^"]*)" endpoint with a request having header "([^"]*)" with value "([^"]*)" should return same header with value "([^"]*)"$`,
		scenario.shouldOverwriteHeaderValue,
	)
	ctx.Step(`^JwtMutatorsOverwrite: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) shouldOverwriteHeaderValue(endpoint, headerName, requestValue, responseValue string) error {
	requestHeaders := map[string]string{headerName: requestValue}

	expectedInBody := []string{fmt.Sprintf(`"%s": "%s"`, headerName, responseValue)}
	asserter := &helpers.BodyContainsPredicate{Expected: expectedInBody}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}

	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, "JWT", asserter, requestHeaders, &tokenFrom)
}

func (s *scenario) shouldOverwriteCookieValue(endpoint, requestValue, responseValue string) error {
	requestHeaders := map[string]string{"Cookie": requestValue}

	expectedInBody := []string{fmt.Sprintf(`"%s": "%s"`, "Cookie", responseValue)}
	asserter := &helpers.BodyContainsPredicate{Expected: expectedInBody}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}

	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, "JWT", asserter, requestHeaders, &tokenFrom)
}
