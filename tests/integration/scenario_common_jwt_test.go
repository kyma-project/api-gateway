package api_gateway

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
)

const apiRuleManifestFile string = "istio-jwt-common.yaml"

func initCommon(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource(apiRuleManifestFile, "istio-jwt")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Common: There is an endpoint secured with JWT on path "([^"]*)"$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", s.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *istioJwtManifestScenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", s.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (s *istioJwtManifestScenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(path, tokenType string, lower, higher int) error {
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, lower, higher)
}
