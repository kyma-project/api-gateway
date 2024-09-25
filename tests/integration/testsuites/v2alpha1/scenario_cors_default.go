package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initDefaultCors(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("cors-default.yaml", "cors-default")

	ctx.Step(`^DefaultCORS: There is an httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^DefaultCORS: The APIRule without CORS set up is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^DefaultCORS: Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and no response header "([^"]*)"$`, scenario.preflightEndpointCallNoResponseHeader)
	ctx.Step(`^DefaultCORS: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
