package istiojwt

import (
	"github.com/cucumber/godog"
)

func initRequiredScopes(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-scopes.yaml", "istio-jwt-scopes")

	ctx.Step(`^Scopes: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Scopes: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Scopes: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with "([^"]*)" "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidToken)
	ctx.Step(`^Scopes: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
