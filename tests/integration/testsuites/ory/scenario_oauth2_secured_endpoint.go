package ory

import (
	_ "embed"

	"github.com/cucumber/godog"
)

func initOAuth2Endpoint(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("oauth-strategy.yaml", "oauth2-secured")

	ctx.Step(`^OAuth2: There is an endpoint secured with OAuth2 introspection$`, scenario.thereIsAHttpbinServiceAndApiRuleIsApplied)
	ctx.Step(
		`^OAuth2: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^OAuth2: Calling the "([^"]*)" endpoint with a invalid token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^OAuth2: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^OAuth2: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
