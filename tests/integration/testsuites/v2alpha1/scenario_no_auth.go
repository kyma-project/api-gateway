package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initNoAuthWildcard(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("no-auth-wildcard.yaml", "no-auth-wildcard")

	ctx.Step(`^Wildcard: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Wildcard: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Wildcard: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Wildcard: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
