package istiojwt

import (
	"github.com/cucumber/godog"
)

func initExposeMethodsOnPathsOAuth2Handler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("paths-and-methods-oauth2-handler.yaml", "paths-and-methods-oauth2")

	ctx.Step(`^ExposeMethodsOnPathsOAuth2Handler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeMethodsOnPathsOAuth2Handler: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^ExposeMethodsOnPathsOAuth2Handler: Calling the "([^"]*)" endpoint with "([^"]*)" method with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithMethodWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^ExposeMethodsOnPathsOAuth2Handler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
