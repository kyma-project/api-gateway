package v2alpha1

import "github.com/cucumber/godog"

func initExtAuthCommon(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("ext-auth-common.yaml", "ext-auth-common")

	ctx.Step(`ExtAuthCommon: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`ExtAuthCommon: There is an endpoint secured with ExtAuth "([^"]*)" on path "([^"]*)"$`, scenario.thereIsAnEndpointWithExtAuth)
	ctx.Step(`ExtAuthCommon: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`ExtAuthCommon: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithHeader)
}

func initExtAuthJwt(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("ext-auth-jwt.yaml", "ext-auth-jwt")

	ctx.Step(`ExtAuthJwt: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`ExtAuthJwt: There is an endpoint secured with ExtAuth "([^"]*)" on path "([^"]*)"$`, scenario.thereIsAnEndpointWithExtAuth)
	ctx.Step(`ExtAuthJwt: The endpoint has JWT restrictions$`, scenario.theEndpointHasJwtRestrictionsWithScope)
	ctx.Step(`ExtAuthJwt: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`ExtAuthJwt: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and no token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithHeaderAndNoToken)
	ctx.Step(`ExtAuthJwt: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and an invalid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithHeaderAndInvalidJwt)
	ctx.Step(`ExtAuthJwt: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithHeaderAndValidJwt)
}
