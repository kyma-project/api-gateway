package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initShortHost(ctx *godog.ScenarioContext, ts *testsuite) {
	scenarioSuccess := ts.createScenario("short-host.yaml", "short-host")

	ctx.Step(`^ShortHost: There is a httpbin service$`, scenarioSuccess.thereIsAHttpbinService)
	ctx.Step(`^ShortHost: The APIRule is applied$`, scenarioSuccess.theAPIRuleIsApplied)
	ctx.Step(`^ShortHost: Calling short host "([^"]*)" with path "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenarioSuccess.callingShortHostWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^ShortHost: Teardown httpbin service$`, scenarioSuccess.teardownHttpbinService)

	scenarioError := ts.createScenario("short-host.yaml", "short-host-error")

	ctx.Step(`^ShortHostError: There is a httpbin service$`, scenarioError.thereIsAHttpbinService)
	ctx.Step(`^ShortHostError: Specifies custom Gateway "([^"]*)"/"([^"]*)"`, scenarioError.specifiesCustomGateway)
	ctx.Step(`^ShortHostError: The APIRule is applied and contains error status with "([^"]*)" message$`, scenarioError.theAPIRuleIsAppliedExpectError)
	ctx.Step(`^ShortHostError: Calling short host "([^"]*)" with path "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenarioError.callingShortHostWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^ShortHostError: Teardown httpbin service$`, scenarioError.teardownHttpbinService)
}
