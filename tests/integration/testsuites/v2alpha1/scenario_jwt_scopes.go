package v2alpha1

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func initJwtScopes(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-scopes.yaml", "jwt-scopes")

	ctx.Step(`^JwtScopes: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtScopes: There is an endpoint secured with JWT on path "([^"]*)" requiring scopes '(\[.*\])'$`, scenario.thereIsAnEndpointWithRequiredScopes)
	ctx.Step(`^JwtScopes: The APIRule is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^JwtScopes: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with "([^"]*)" "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidToken)
	ctx.Step(`^JwtScopes: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithRequiredScopes(path string, scopes string) error {
	s.ManifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "RequiredScopes")] = scopes
	return nil
}
