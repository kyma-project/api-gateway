package ory

import (
	"github.com/cucumber/godog"
)

func initV1Beta1AllowDelete(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta1-allow-delete.yaml", "v1beta1-allow-delete")

	ctx.Step(`^v1beta1AllowDelete: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^v1beta1AllowDelete: The APIRule is deleted using v1beta2$`, scenario.theV1beta2APIRuleIsDeleted)
	ctx.Step(`^v1beta1AllowDelete: APIRule is not found$`, scenario.theAPIRuleIsNotFound)
}
