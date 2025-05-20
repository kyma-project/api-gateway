package ory

import (
	_ "embed"

	"github.com/cucumber/godog"
)

func initUnsecured(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("unsecured.yaml", "unsecured")

	ctx.Step(`^Unsecured: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Unsecured: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^Unsecured: Calling the "([^"]*)" endpoint with any token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^Unsecured: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^Unsecured: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
