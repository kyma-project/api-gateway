package istiojwt

import (
	"github.com/cucumber/godog"
)

func initV1Beta2IstioJWT(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta2-istio-jwt.yaml", "v1beta2-istio-jwt")

	ctx.Step(`v1beta2IstioJWT: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
}

func initV1Beta2NoAuthHandler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("v1beta2-no-auth-handler.yaml", "v1beta2-no-auth-handler")

	ctx.Step(`^v1beta2NoAuthHandler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^v1beta2NoAuthHandler: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^v1beta2NoAuthHandler: Calling the "([^"]*)" endpoint with "([^"]*)" method with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^v1beta2NoAuthHandler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
