package api_gateway

import (
	"github.com/cucumber/godog"
)

func initTokenFrom(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-token-from.yaml", "istio-jwt-token-from")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`JwtTokenFrom: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtTokenFrom: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtTokenFrom: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`JwtTokenFrom: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from header "([^"]*)" and prefix "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween)
	ctx.Step(`JwtTokenFrom: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from parameter "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween)
}
