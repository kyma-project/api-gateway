package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initRequestHeaders(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-request-header.yaml", "jwt-request-header")

	ctx.Step(`MutatorHeader: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`MutatorHeader: There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)"$`, scenario.thereIsAnEndpointWithHeader)
	ctx.Step(`MutatorHeader: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`MutatorHeader: Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.callinTheEndpointShouldResultInBodyContaining)
	ctx.Step(`MutatorHeader: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initRequestCookies(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-request-cookie.yaml", "jwt-request-cookie")

	ctx.Step(`MutatorCookie: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`MutatorCookie: There is an endpoint on path "([^"]*)" with a cookie mutator setting "([^"]*)" cookie to "([^"]*)"$`, scenario.thereIsAnEndpointWithCookie)
	ctx.Step(`MutatorCookie: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`MutatorCookie: Calling the "([^"]*)" endpoint should return response with cookie "([^"]*)" with value "([^"]*)"$`, scenario.callinTheEndpointShouldResultInBodyContaining)
	ctx.Step(`MutatorCookie: Teardown httpbin service$`, scenario.teardownHttpbinService)
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
