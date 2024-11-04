package ory

import (
	"github.com/cucumber/godog"
)

func initMigrationNoAuthV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("migration-noauth-v1beta1.yaml", "migration-noauth-v1beta1")

	ctx.Step(`^migrationNoAuthV1beta1: There is a httpbin service with Istio injection enabled$`, scenario.thereIsAHttpbinServiceWithIstioInjection)
	ctx.Step(`^migrationNoAuthV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^migrationNoAuthV1beta1: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^migrationNoAuthV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
	ctx.Step(`^migrationNoAuthV1beta1: Wait for "([^"]*)" seconds$`, scenario.waitForSeconds)
	ctx.Step(`^migrationNoAuthV1beta1: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
}
