package istiojwt

import (
	"fmt"
	"github.com/cucumber/godog"
	"strings"
)

func initJwtAndAllow(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-and-unrestricted.yaml", "istio-jwt-unrestricted")

	ctx.Step(`^JwtAndUnrestricted: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtAndUnrestricted: There is an endpoint secured with JWT on path "([^"]*)"$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^JwtAndUnrestricted: There is an endpoint with handler "([^"]*)" on path "([^"]*)"$`, scenario.thereIsAnEndpointWithHandler)
	ctx.Step(`^JwtAndUnrestricted: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtAndUnrestricted: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtAndUnrestricted: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtAndUnrestricted: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithHandler(handler, handlerPath string) {
	s.ManifestTemplate[fmt.Sprintf("%sEndpoint%s", strings.TrimPrefix(handlerPath, "/"), "Handler")] = handler
}
