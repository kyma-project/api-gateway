package istiojwt

import (
	"github.com/cucumber/godog"
)

func initDefaultCors(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-default-cors.yaml", "istio-default-cors")

	ctx.Step(`^DefaultCORS: There is an httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^DefaultCORS: The APIRule without CORS set up is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^DefaultCORS: Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and response header "([^"]*)" with value "([^"]*)"$`, scenario.preflightEndpointCallResponseHeaders)
	ctx.Step(`^DefaultCORS: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
