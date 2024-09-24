package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initShortHost(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("short-host.yaml", "short-host")

	ctx.Step(`^ShortHost: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ShortHost: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^ShortHost: Calling short host "([^"]*)" with path "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenario.callingShortHostWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^ShortHost: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
