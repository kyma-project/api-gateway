package istiojwt

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

func initCommon(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-common.yaml", "istio-jwt-common")

	ctx.Step(`^Common: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Common: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^Common: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initRegex(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-common.yaml", "istio-jwt-common")

	ctx.Step(`^Regex: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Regex: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^Regex: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Regex: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Regex: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Regex: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^Regex: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func initPrefix(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-common.yaml", "istio-jwt-common")

	ctx.Step(`^Prefix: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Prefix: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^Prefix: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^Prefix: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Prefix: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Prefix: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^Prefix: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	requestHeaders := map[string]string{testcontext.AuthorizationHeaderName: testcontext.AnyToken}
	return s.httpClient.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s%s", s.Url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
