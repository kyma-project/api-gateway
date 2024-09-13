package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initJwtCommon(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-common.yaml", "jwt-common")

	ctx.Step(`^Common: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Common: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^Common: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initJwtWildcard(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-common.yaml", "jwt-wildcard")

	ctx.Step(`^Wildcard: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Wildcard: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^Wildcard: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Wildcard: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Wildcard: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Wildcard: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^Wildcard: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
