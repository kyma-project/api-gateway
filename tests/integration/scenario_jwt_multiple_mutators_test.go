package api_gateway

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
)

func initMultipleMutators(ctx *godog.ScenarioContext) {
	s, err := CreateIstioJwtScenario("istio-jwt-multiple-mutators.yaml", "istio-jwt-multiple-mutators")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`JwtMultipleMutators: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtMultipleMutators: There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)" and "([^"]*)" header to "([^"]*)"$`, scenario.thereIsAnEndpointWithHeaderMutatorWithTwoHeaders)
	ctx.Step(`JwtMultipleMutators: There is an endpoint on path "([^"]*)" with a cookie mutator setting "([^"]*)" cookie to "([^"]*)" and "([^"]*)" cookie to "([^"]*)"$`, scenario.thereIsAnEndpointWithCookieMutatorWithTwoCookies)
	ctx.Step(`JwtMultipleMutators: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtMultipleMutators: Calling the "([^"]*)" endpoint should return response with cookies "([^"]*)" with value "([^"]*)" and "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithKeyValuePairs)
	ctx.Step(`JwtMultipleMutators: Calling the "([^"]*)" endpoint should return response with headers "([^"]*)" with value "([^"]*)" and "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithKeyValuePairs)
	ctx.Step(`JwtMultipleMutators: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithHeaderMutatorWithTwoHeaders(_, header1, header1Value, header2, header2Value string) error {
	s.manifestTemplate["header1"] = header1
	s.manifestTemplate["header1Value"] = header1Value
	s.manifestTemplate["header2"] = header2
	s.manifestTemplate["header2Value"] = header2Value
	return nil
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithCookieMutatorWithTwoCookies(_, cookie1, cookie1Value, cookie2, cookie2Value string) error {
	s.manifestTemplate["cookie1"] = cookie1
	s.manifestTemplate["cookie1Value"] = cookie1Value
	s.manifestTemplate["cookie2"] = cookie2
	s.manifestTemplate["cookie2Value"] = cookie2Value
	return nil
}

func (s *istioJwtManifestScenario) shouldReturnResponseWithKeyValuePairs(endpoint, k1, v1, k2, v2 string) error {
	expectedInBody := []string{
		fmt.Sprintf(`"%s": "%s"`, k1, v1),
		fmt.Sprintf(`"%s": "%s"`, k2, v2),
	}

	asserter := &helpers.BodyContainsPredicate{Expected: expectedInBody}
	tokenFrom := tokenFrom{
		From:     authorizationHeaderName,
		Prefix:   authorizationHeaderPrefix,
		AsHeader: true,
	}

	return callingEndpointWithHeadersWithRetries(fmt.Sprintf("%s/%s", s.url, strings.TrimLeft(endpoint, "/")), "JWT", asserter, nil, &tokenFrom)
}
