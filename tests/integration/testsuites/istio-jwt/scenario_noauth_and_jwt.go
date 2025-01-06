package istiojwt

import (
	"github.com/cucumber/godog"
)

func initJwtAndNoAuth(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-and-no-auth.yaml", "istio-oauth")

	ctx.Step(`^NoAuth_JWT: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^NoAuth_JWT: There is an endpoint secured with JWT on path "([^"]*)" requiring scopes '(\[.*\])'$`, scenario.thereIsAnEndpointWithRequiredScopes)
	ctx.Step(`^NoAuth_JWT: There is an endpoint secured with NoAuth on path "([^"]*)"`, func() error { return nil })
	ctx.Step(`^NoAuth_JWT: The APIRules are applied$`, scenario.theAPIRulesAreApplied)
	ctx.Step(`^NoAuth_JWT: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with "([^"]*)" "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAValidToken)
	ctx.Step(`^NoAuth_JWT: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^NoAuth_JWT: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^NoAuth_JWT: Calling the "([^"]*)" endpoint on prefix "([^"]*)" without a token should result in status between (\d+) and (\d+)$`, scenario.callingPrefixWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^NoAuth_JWT: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
