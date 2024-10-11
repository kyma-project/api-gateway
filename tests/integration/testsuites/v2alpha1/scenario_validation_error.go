package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initValidationError(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("validation-error.yaml", "validation-error")

	ctx.Step(`ValidationError: APIRule is applied and contains error status with "([^"]*)" message$`, scenario.theAPIRuleV2Alpha1IsAppliedExpectError)
}
