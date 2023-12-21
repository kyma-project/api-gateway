package ory

import (
	"github.com/cucumber/godog"
)

func initExposeMethodsOnPathsAllowHandler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("multiple-paths-and-methods-allow-handler.yaml", "multiple-paths-and-methods-allow-handler")

	ctx.Step(`^ExposeMethodsOnPathsAllowHandler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeMethodsOnPathsAllowHandler: The APIRule is applied$`, scenario.theManifestIsApplied)
	ctx.Step(`^ExposeMethodsOnPathsAllowHandler: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^ExposeMethodsOnPathsAllowHandler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
