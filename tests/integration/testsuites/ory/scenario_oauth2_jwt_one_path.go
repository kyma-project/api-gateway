package ory

import (
	_ "embed"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/auth"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
)

func initOAuth2JWTOnePath(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-oauth-one-path-strategy.yaml", "oauth2-jwt-one-path")

	ctx.Step(`^OAuth2JWT1Path: There is an deployment secured with both JWT and OAuth2 introspection on path /image$`, scenario.thereIsAHttpbinServiceAndApiRuleIsApplied)
	ctx.Step(
		`^OAuth2JWT1Path: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^OAuth2JWT1Path: Calling the "([^"]*)" endpoint with a invalid token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^OAuth2JWT1Path: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^OAuth2JWT1Path: Calling the "([^"]*)" endpoint with a valid OAuth2 token in token from header "([^"]*)" should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidOauthTokenInTokenFromHeader,
	)
	ctx.Step(`^OAuth2JWT1Path: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) callingTheEndpointWithValidOauthTokenInTokenFromHeader(path string, tokenHeader string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}

	token, err := auth.GetAccessToken(*s.oauth2Cfg, "")
	if err != nil {
		return err
	}

	requestHeaders := make(map[string]string)
	requestHeaders[tokenHeader] = token

	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s%s", s.Url, path), asserter)
}
