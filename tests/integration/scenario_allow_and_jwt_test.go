package api_gateway

import (
	"github.com/cucumber/godog"
)

func initJwtAndAllow(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-and-allow.yaml", "istio-jwt")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Common: There is an endpoint secured with JWT on path "([^"]*)"$ and an unrestricted path "([^"]*)"$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
}
