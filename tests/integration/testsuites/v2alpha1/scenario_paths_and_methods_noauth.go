package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initExposeMethodsOnPathsNoAuthHandler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("paths-and-methods-noauth.yaml", "paths-and-methods-noauth")

	ctx.Step(`^ExposeMethodsOnPathsNoAuthHandler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeMethodsOnPathsNoAuthHandler: The APIRule is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^ExposeMethodsOnPathsNoAuthHandler: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^ExposeMethodsOnPathsNoAuthHandler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
