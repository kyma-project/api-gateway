package ory

import (
	"github.com/cucumber/godog"
)

func initMigrationAllowV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("migration-allow-v1beta1.yaml", "migration-allow-v1beta1")

	ctx.Step(`^allowMigrationV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^allowMigrationV1beta1: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^allowMigrationV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
}
