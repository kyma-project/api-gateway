package istiojwt

import (
	"github.com/cucumber/godog"
)

func initAudience(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-audiences.yaml", "istio-jwt-audiences")

	ctx.Step(`^Audiences: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Audiences: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Audiences: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with "([^"]*)" "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidToken)
	ctx.Step(`^Audiences: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
