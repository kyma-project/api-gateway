package api_gateway

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func initAudience(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-audiences.yaml", "istio-jwt-audiences")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Audiences: There is an endpoint secured with JWT on path "([^"]*)" requiring audiences '(\[.*\])'$`, scenario.thereIsAnEndpointWithAudiences)
	ctx.Step(`Audiences: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Audiences: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with audiences "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidToken)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithAudiences(path string, audiences string) error {
	s.manifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "Audiences")] = audiences
	return nil
}
