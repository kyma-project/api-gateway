package ory

import (
	_ "embed"
	"github.com/cucumber/godog"
)

func initScenarioServicePerPath(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("service-per-path.yaml", "service-per-path")

	ctx.Step(`^Service per path: There is a httpbin service`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Service per path: There is a helloworld service and an APIRule with two endpoints exposed with different services, one on spec level and one on rule level$`, scenario.theManifestIsApplied)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Service per path: Calling the endpoint "([^"]*)" and "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointsWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Service per path: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
