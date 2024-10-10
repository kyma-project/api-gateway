package ory

import (
	"github.com/cucumber/godog"
)

func initMigrationOauth2IntrospectionJwtV1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("migration-oauth2-introspection-v1beta1.yaml", "migration-oauth2-introspection-v1beta1")

	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: There is a httpbin service with Istio injection enabled$`, scenario.thereIsAHttpbinServiceWithIstioInjection)
	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: VirtualService owned by APIRule has httpbin service as destination$`, scenario.thereIsApiRuleVirtualServiceWithHttpbinServiceDestination)
	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "([^"]*)" owned by APIRule does not exist$`, scenario.resourceOwnedByApiRuleDoesNotExist)
	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: Resource of Kind "([^"]*)" owned by APIRule exists$`, scenario.resourceOwnedByApiRuleExists)
	ctx.Step(`^migrationOAuth2IntrospectionJwtV1beta1: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)

}
