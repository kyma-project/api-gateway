package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initExposeAsterisk(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("asterisk-paths.yaml", "asterisk")

	ctx.Step(`^ExposeAsterisk: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeAsterisk: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^ExposeAsterisk: Calling the "([^"]*)" endpoint with "([^"]*)" method should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodShouldResultInStatusBetween)
	ctx.Step(`^ExposeAsterisk: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
