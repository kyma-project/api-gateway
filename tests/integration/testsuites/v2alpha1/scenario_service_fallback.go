package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initServiceFallback(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("service-fallback.yaml", "jwt-service-fallback")

	ctx.Step(`^ServiceFallback: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ServiceFallback: There is an endpoint secured with JWT on path "([^"]*)" with service definition$`, scenario.thereIsAnEndpointWithServiceDefinition)
	ctx.Step(`^ServiceFallback: There is an endpoint secured with JWT on path "([^"]*)"$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^ServiceFallback: The APIRule with service on root level is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^ServiceFallback: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^ServiceFallback: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithServiceDefinition(path string) {
	s.ManifestTemplate["jwtSecuredPathWithService"] = path
}
