package api_gateway

import (
	"github.com/cucumber/godog"
)

func initTokenFromHeaders(ctx *godog.ScenarioContext) {
	jwtHeaderName = "x-jwt-token"

	s, err := CreateScenarioWithRawAPIResource("istio-jwt-token-from-headers.yaml", "istio-jwt-token-from-headers")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`JwtTokenFromHeaders: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtTokenFromHeaders: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtTokenFromHeaders: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`JwtTokenFromHeaders: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from header "([^"]*)" and prefix "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween)
	ctx.Step(`JwtTokenFromHeaders: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initTokenFromParams(ctx *godog.ScenarioContext) {
	fromParamName = "jwt_token"

	s, err := CreateScenarioWithRawAPIResource("istio-jwt-token-from-params.yaml", "istio-jwt-token-from-params")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`JwtTokenFromParams: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtTokenFromParams: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtTokenFromParams: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`JwtTokenFromParams: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from parameter "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween)
	ctx.Step(`JwtTokenFromParams: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
