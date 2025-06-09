package istiojwt

import (
	"github.com/cucumber/godog"
)

func initJwtIssuerJwksNotMatch(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-issuer-jwks-not-match.yaml", "jwt-issuer-jwks-not-match")

	ctx.Step(`^JwtIssuerJwksNotMatch: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtIssuerJwksNotMatch: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtIssuerJwksNotMatch: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in body containing "([^"]*)"$`, scenario.callingTheEndpointWithValidTokenShouldResultInBodyContaining)
	ctx.Step(`^JwtIssuerJwksNotMatch: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
