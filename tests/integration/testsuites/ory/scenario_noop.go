package ory

import (
	"github.com/cucumber/godog"
)

func initNoop(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("noop.yaml", "noop")

	ctx.Step(`^Noop: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Noop: The APIRule is applied$`, scenario.theManifestIsApplied)
	ctx.Step(`^Noop: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Noop: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
