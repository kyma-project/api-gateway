package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initValidationError(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("validation-error.yaml", "validation-error")

	ctx.Step(`ValidationError: The misconfigured APIRule is applied$`, scenario.theMisconfiguredAPIRuleIsApplied)
	ctx.Step(`ValidationError: APIRule has status "([^"]*)" with description containing "([^"]*)"$`, scenario.theAPIRuleHasStatusWithDesc)
}
