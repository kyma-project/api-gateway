package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initNoAuthWildcard(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("no-auth-wildcard.yaml", "no-auth-wildcard")

	ctx.Step(`^NoAuthWildcard: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^NoAuthWildcard: The APIRule is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^NoAuthWildcard: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^NoAuthWildcard: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
