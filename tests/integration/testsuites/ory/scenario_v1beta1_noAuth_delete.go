package ory

import (
	"github.com/cucumber/godog"
)

func initV1Beta1NoAuthDelete(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta1-noauth-delete.yaml", "v1beta1-noauth-delete")

	ctx.Step(`^v1beta1NoAuthDelete: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^v1beta1NoAuthDelete: The APIRule is deleted using v1beta2$`, scenario.theV1beta2APIRuleIsDeleted)
	ctx.Step(`^v1beta1NoAuthDelete: APIRule is not found$`, scenario.theAPIRuleIsNotFound)
}
