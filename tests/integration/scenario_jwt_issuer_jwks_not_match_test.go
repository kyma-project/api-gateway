package api_gateway

import (
	"github.com/cucumber/godog"
)

func initJwtIssuerJwksNotMatch(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-issuer-jwks-not-match.yaml", "jwt-issuer-jwks-not-match")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`JwtIssuerJwksNotMatch: There is an endpoint secured with JWT on path "([^"]*)" with invalid issuer and jwks$`, scenario.emptyStep)
	ctx.Step(`JwtIssuerJwksNotMatch: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtIssuerJwksNotMatch: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in body containing "([^"]*)"$`, scenario.callingTheEndpointWithValidTokenShouldResultInBodyContaining)
}
