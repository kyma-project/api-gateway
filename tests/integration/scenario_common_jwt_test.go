package api_gateway

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
)

func initCommon(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-common.yaml", "istio-jwt-common")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Common: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`Common: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`Common: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`Common: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	requestHeaders := map[string]string{authorizationHeaderName: anyToken}
	return helper.CallEndpointWithHeadersWithRetries(requestHeaders, fmt.Sprintf("%s%s", s.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
