package ory

import (
	"github.com/cucumber/godog"
)

func initExposeMethodsOnPathsJwtHandler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("paths-and-methods-jwt-handler.yaml", "paths-and-methods-jwt")

	ctx.Step(`^ExposeMethodsOnPathsJwtHandler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeMethodsOnPathsJwtHandler: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^ExposeMethodsOnPathsJwtHandler: Calling the "([^"]*)" endpoint with "([^"]*)" method with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithMethodWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^ExposeMethodsOnPathsJwtHandler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
