package api_gateway

import (
	"github.com/cucumber/godog"
)

func initJwtAndOauth(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-and-oauth.yaml", "istio-oauth")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`OAuth2: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`OAuth2: There is an endpoint secured with JWT on path "([^"]*)" requiring scopes '(\[.*\])'$`, scenario.thereIsAnEndpointWithRequiredScopes)
	ctx.Step(`OAuth2: There is an endpoint secured with OAuth2 on path "([^"]*)" requiring scopes '(\[.*\])'$`, scenario.thereIsAnEndpointWithRequiredScopes)
	ctx.Step(`OAuth2: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`OAuth2: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`OAuth2: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`OAuth2: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
}
