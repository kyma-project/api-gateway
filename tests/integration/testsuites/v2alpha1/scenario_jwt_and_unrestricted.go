package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initJwtAndAllow(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-and-unrestricted.yaml", "jwt-unrestricted")

	ctx.Step(`^JwtAndUnrestricted: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtAndUnrestricted: There is an endpoint secured with JWT on path "([^"]*)" and /headers endpoint exposed with noAuth$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^JwtAndUnrestricted: The APIRule is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^JwtAndUnrestricted: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtAndUnrestricted: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtAndUnrestricted: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
