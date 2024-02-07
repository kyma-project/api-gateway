package istiojwt

import (
	"github.com/cucumber/godog"
)

func initExposeMethodsOnPathsAllowMethodsHandler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("paths-and-methods-allow-methods-handler.yaml", "paths-and-methods-allow-methods")

	ctx.Step(`^ExposeMethodsOnPathsAllowMethodsHandler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeMethodsOnPathsAllowMethodsHandler: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^ExposeMethodsOnPathsAllowMethodsHandler: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^ExposeMethodsOnPathsAllowMethodsHandler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
