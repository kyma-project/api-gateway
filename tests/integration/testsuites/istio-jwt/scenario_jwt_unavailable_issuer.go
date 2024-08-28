package istiojwt

import (
	"github.com/cucumber/godog"
)

func initJwtUnavailableIssuer(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-unavailable-issuer.yaml", "jwt-unavailable-issuer")

	ctx.Step(`^JwtIssuerUnavailable: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtIssuerUnavailable: There is an endpoint secured with JWT on path "([^"]*)" with invalid issuer and jwks$`, scenario.emptyStep)
	ctx.Step(`^JwtIssuerUnavailable: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtIssuerUnavailable: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in body containing "([^"]*)"$`, scenario.callingTheEndpointWithValidTokenShouldResultInBodyContaining)
	ctx.Step(`^JwtIssuerUnavailable: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
