package ory

import (
	"github.com/cucumber/godog"
)

func initDeleteNoAuthV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("delete-noauth-v1beta1.yaml", "delete-noauth-v1beta1")

	ctx.Step(`^deleteNoAuthV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^deleteNoAuthV1beta1: The APIRule is deleted using v1beta2$`, scenario.theAPIRuleIsDeletedUsingV1beta2Version)
	ctx.Step(`^deleteNoAuthV1beta1: APIRule is not found$`, scenario.theAPIRuleIsNotFound)
}
