package istiojwt

import (
	"github.com/cucumber/godog"
)

func initExposeMethodsOnPathsNoopHandler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("paths-and-methods-noop-handler.yaml", "paths-and-methods-noop")

	ctx.Step(`^ExposeMethodsOnPathsNoopHandler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeMethodsOnPathsNoopHandler: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^ExposeMethodsOnPathsNoopHandler: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^ExposeMethodsOnPathsNoopHandler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
