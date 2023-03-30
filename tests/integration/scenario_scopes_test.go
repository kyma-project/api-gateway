package api_gateway

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func initRequiredScopes(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-scopes.yaml", "istio-jwt-scopes")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Scopes: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`Scopes: There is an endpoint secured with JWT on path "([^"]*)" requiring scopes '(\[.*\])'$`, scenario.thereIsAnEndpointWithRequiredScopes)
	ctx.Step(`Scopes: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Scopes: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with scope claims "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidToken)
	ctx.Step(`Scopes: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithRequiredScopes(path string, scopes string) error {
	s.manifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "RequiredScopes")] = scopes
	return nil
}
