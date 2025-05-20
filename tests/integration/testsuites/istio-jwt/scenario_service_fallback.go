package istiojwt

import (
	"github.com/cucumber/godog"
)

func initJwtServiceFallback(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-service-fallback.yaml", "istio-jwt-service-fallback")

	ctx.Step(`^ServiceFallback: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ServiceFallback: There is an endpoint secured with JWT on path "([^"]*)" with service definition$`, scenario.thereIsAnEndpointWithServiceDefinition)
	ctx.Step(`^ServiceFallback: There is an endpoint secured with JWT on path "([^"]*)"$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^ServiceFallback: The APIRule with service on root level is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^ServiceFallback: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^ServiceFallback: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithServiceDefinition(path string) {
	s.ManifestTemplate["jwtSecuredPathWithService"] = path
}
