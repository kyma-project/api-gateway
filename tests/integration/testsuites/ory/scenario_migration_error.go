package ory

import (
	"github.com/cucumber/godog"
)

func initMigrationError(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("migration-noop-v1beta1.yaml", "migration-error")

	ctx.Step(`^migrationError: There is a httpbin service with Istio injection enabled$`, scenario.thereIsAHttpbinServiceWithIstioInjection)
	ctx.Step(`^migrationError: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^migrationError: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^migrationError: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
	ctx.Step(`^migrationError: Resource of Kind "([^"]*)" owned by APIRule does not exist$`, scenario.resourceOwnedByApiRuleDoesNotExist)
	ctx.Step(`^migrationError: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^migrationError: Resource of Kind "([^"]*)" owned by APIRule exists$`, scenario.resourceOwnedByApiRuleExists)
	ctx.Step(`^migrationError: The APIRule contains original-version annotation set to "([^"]*)"$`, scenario.apiRuleContainsOriginalVersionAnnotation)
}
