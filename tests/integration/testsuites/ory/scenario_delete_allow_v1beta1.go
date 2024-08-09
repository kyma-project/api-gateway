package ory

import (
	"github.com/cucumber/godog"
)

func initDeleteAllowV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("delete-allow-v1beta1.yaml", "delete-allow-v1beta1")

	ctx.Step(`^deleteAllowV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^deleteAllowV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
	ctx.Step(`^deleteAllowV1beta1: The APIRule is deleted using v2alpha1$`, scenario.theAPIRuleIsDeletedUsingv2alpha1Version)
	ctx.Step(`^deleteAllowV1beta1: APIRule is not found$`, scenario.theAPIRuleIsNotFound)
}
