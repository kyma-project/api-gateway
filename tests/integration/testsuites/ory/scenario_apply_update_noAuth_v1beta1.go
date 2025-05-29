package ory

import (
	"github.com/cucumber/godog"
)

func initApplyUpdatev1beta1(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta1-noop.yaml", "v1beta1-noop")

	ctx.Step(`^applyUpdateNoAuthV1beta1: There is a httpbin service with Istio injection enabled$`, scenario.thereIsAHttpbinServiceWithIstioInjection)
	ctx.Step(`^applyUpdateNoAuthV1beta1: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^applyUpdateNoAuthV1beta1: APIRule has status "([^"]*)"$`, scenario.theAPIRuleHasStatus)
	ctx.Step(`^applyUpdateNoAuthV1beta1: The APIRule contains original-version annotation set to "([^"]*)"$`, scenario.apiRuleContainsOriginalVersionAnnotation)
	ctx.Step(`^applyUpdateNoAuthV1beta1: Resource of Kind "([^"]*)" owned by APIRule exists$`, scenario.resourceOwnedByApiRuleExists)
	ctx.Step(`^applyUpdateNoAuthV1beta1: The APIRule is updated using manifest "([^"]*)"$`, scenario.theAPIRuleIsUpdated)
	ctx.Step(`^applyUpdateNoAuthV1beta1: Teardown httpbin service$`, scenario.teardownHttpbinService)
	ctx.Step(`^applyUpdateNoAuthV1beta1: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
}
