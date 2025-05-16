package ory

import (
	"github.com/cucumber/godog"
)

func initMigrationNoopV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("migration-noop-v1beta1.yaml", "migration-noop-v1beta1")

	ctx.Step(`^migrationNoopV1beta1: There is a httpbin service with Istio injection enabled$`, scenario.thereIsAHttpbinServiceWithIstioInjection)
	ctx.Step(`^migrationNoopV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^migrationNoopV1beta1: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^migrationNoopV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
	ctx.Step(`^migrationNoopV1beta1: Resource of Kind "([^"]*)" owned by APIRule does not exist$`, scenario.resourceOwnedByApiRuleDoesNotExist)
	ctx.Step(`^migrationNoopV1beta1: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^migrationNoopV1beta1: Resource of Kind "([^"]*)" owned by APIRule exists$`, scenario.resourceOwnedByApiRuleExists)
}
