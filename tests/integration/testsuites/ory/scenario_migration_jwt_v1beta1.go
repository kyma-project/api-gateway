package ory

import (
	"github.com/cucumber/godog"
)

func initMigrationJwtV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("migration-jwt-v1beta1.yaml", "migration-jwt-v1beta1")

	ctx.Step(`^migrationJwtV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^migrationJwtV1beta1: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^migrationJwtV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
}
