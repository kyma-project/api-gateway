package istiojwt

import (
	"github.com/cucumber/godog"
)

func initAllow(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("allow.yaml", "allow")

	ctx.Step(`^Allow: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Allow: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Allow: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Allow: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
