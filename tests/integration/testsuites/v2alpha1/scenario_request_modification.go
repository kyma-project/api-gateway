package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initRequestHeaders(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-request-header.yaml", "jwt-request-header")

	ctx.Step(`^RequestHeader: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^RequestHeader: There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)"$`, scenario.thereIsAnEndpointWithHeader)
	ctx.Step(`^RequestHeader: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^RequestHeader: Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.callingTheEndpointShouldResultInBodyContaining)
	ctx.Step(`^RequestHeader: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initRequestCookies(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-request-cookie.yaml", "jwt-request-cookie")

	ctx.Step(`^RequestCookie: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^RequestCookie: There is an endpoint on path "([^"]*)" with a cookie mutator setting "([^"]*)" cookie to "([^"]*)"$`, scenario.thereIsAnEndpointWithCookie)
	ctx.Step(`^RequestCookie: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^RequestCookie: Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.callingTheEndpointShouldResultInBodyContaining)
	ctx.Step(`^RequestCookie: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithHeader(_, header, headerValue string) error {
	s.ManifestTemplate["header"] = header
	s.ManifestTemplate["headerValue"] = headerValue
	return nil
}

func (s *scenario) thereIsAnEndpointWithCookie(_, cookie, cookieValue string) error {
	s.ManifestTemplate["cookie"] = cookie
	s.ManifestTemplate["cookieValue"] = cookieValue
	return nil
}
