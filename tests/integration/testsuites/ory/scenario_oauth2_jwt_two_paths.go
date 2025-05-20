package ory

import (
	_ "embed"

	"github.com/cucumber/godog"
)

func initOAuth2JWTTwoPaths(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-oauth-strategy.yaml", "oauth2-jwt-two-paths")

	ctx.Step(`^OAuth2JWTTwoPaths: There is a deployment secured with OAuth2 on path /headers and JWT on path /image$`, scenario.thereIsAHttpbinServiceAndApiRuleIsApplied)
	ctx.Step(
		`^OAuth2JWTTwoPaths: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^OAuth2JWTTwoPaths: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^OAuth2JWTTwoPaths: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^OAuth2JWTTwoPaths: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
