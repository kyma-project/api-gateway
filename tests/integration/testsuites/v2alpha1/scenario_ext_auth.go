package v2alpha1

import "github.com/cucumber/godog"

func initExtAuthCommon(ctx *godog.ScenarioContext, ts *testsuite) {
	s := ts.createScenario("ext-auth-common.yaml", "ext-auth-common")

	ctx.Step(`^ExtAuthCommon: There is a httpbin service$`, s.thereIsAHttpbinService)
	ctx.Step(`^ExtAuthCommon: There is an endpoint secured with ExtAuth "([^"]*)" on path "([^"]*)"$`, s.thereIsAnEndpointWithExtAuth)
	ctx.Step(`^ExtAuthCommon: The APIRule is applied$`, s.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^ExtAuthCommon: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" should result in status between (\d+) and (\d+)$`, s.callingTheEndpointWithHeader)
}

func initExtAuthJwt(ctx *godog.ScenarioContext, ts *testsuite) {
	s := ts.createScenario("ext-auth-jwt.yaml", "ext-auth-jwt")

	ctx.Step(`^ExtAuthJwt: There is a httpbin service$`, s.thereIsAHttpbinService)
	ctx.Step(`^ExtAuthJwt: There is an endpoint secured with ExtAuth "([^"]*)" on path "([^"]*)"$`, s.thereIsAnEndpointWithExtAuth)
	ctx.Step(`^ExtAuthJwt: The endpoint has JWT restrictions$`, s.theEndpointHasJwtRestrictionsWithScope)
	ctx.Step(`^ExtAuthJwt: The APIRule is applied$`, s.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^ExtAuthJwt: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and no token should result in status between (\d+) and (\d+)$`, s.callingTheEndpointWithHeader)
	ctx.Step(`^ExtAuthJwt: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and an invalid "([^"]*)" token should result in status between (\d+) and (\d+)$`, s.callingTheEndpointWithHeaderAndInvalidJwt)
	ctx.Step(`^ExtAuthJwt: Calling the "([^"]*)" endpoint with header "([^"]*)" with value "([^"]*)" and a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, s.callingTheEndpointWithHeaderAndValidJwt)
}
