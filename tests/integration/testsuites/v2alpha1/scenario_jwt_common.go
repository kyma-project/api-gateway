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

func initJwtRegex(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-common.yaml", "jwt-regex")

	ctx.Step(`^Regex: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Regex: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^Regex: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Regex: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Regex: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Regex: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^Regex: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initJwtPrefix(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-common.yaml", "jwt-prefix")

	ctx.Step(`^Prefix: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Prefix: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^Prefix: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Prefix: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Prefix: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Prefix: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^Prefix: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
