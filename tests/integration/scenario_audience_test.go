package api_gateway

import (
	"fmt"
	"github.com/cucumber/godog"
	"strings"
)

const audiencesManifestFile string = "istio-jwt-audiences.yaml"

func initAudience(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource(audiencesManifestFile, "istio-jwt-audiences")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Audiences: There is an endpoint secured with JWT on path "([^"]*)" requiring audiences '(\[.*\])'$`, scenario.thereIsAnEndpointWithAudiences)
	ctx.Step(`Audiences: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Audiences: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with audiences "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidJWTToken)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithAudiences(path string, audiences string) error {
	s.manifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "Audiences")] = audiences
	return nil
}
