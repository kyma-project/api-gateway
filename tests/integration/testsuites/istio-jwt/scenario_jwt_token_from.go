package istiojwt

import (
	"github.com/cucumber/godog"
)

func initTokenFromHeaders(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-token-from-headers.yaml", "istio-jwt-token-from-headers")
	scenario.ManifestTemplate["JWTHeaderName"] = "x-jwt-token"

	ctx.Step(`^JwtTokenFromHeaders: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtTokenFromHeaders: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtTokenFromHeaders: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromHeaders: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from default header should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromHeaders: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from header "([^"]*)" and prefix "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromHeaders: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initTokenFromParams(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-token-from-params.yaml", "istio-jwt-token-from-params")
	scenario.ManifestTemplate["FromParamName"] = "jwt_token"

	ctx.Step(`^JwtTokenFromParams: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtTokenFromParams: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtTokenFromParams: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromParams: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from default header should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromParams: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from parameter "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromParams: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
