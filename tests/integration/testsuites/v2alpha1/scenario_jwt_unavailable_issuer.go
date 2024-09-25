package v2alpha1

import (
	"github.com/cucumber/godog"
)

func initJwtUnavailableIssuer(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-unavailable-issuer.yaml", "jwt-unavailable-issuer")

	ctx.Step(`^JwtUnavailableIssuer: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtUnavailableIssuer: There is an endpoint secured with JWT on path "([^"]*)" with invalid issuer and jwks$`, scenario.emptyStep)
	ctx.Step(`^JwtUnavailableIssuer: The APIRule is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^JwtUnavailableIssuer: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in body containing "([^"]*)"$`, scenario.callingTheEndpointWithValidTokenShouldResultInBodyContaining)
	ctx.Step(`^JwtUnavailableIssuer: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
