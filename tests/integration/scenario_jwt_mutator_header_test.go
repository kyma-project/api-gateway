package api_gateway

import (
	"fmt"
	"github.com/cucumber/godog"
)

func initMutatorHeader(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-mutator-header.yaml", "istio-jwt-mutator-header")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`JwtMutatorHeader: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtMutatorHeader: There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)"$`, scenario.thereIsAnEndpointWithHeaderMutator)
	ctx.Step(`JwtMutatorHeader: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtMutatorHeader: Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithHeader)
	ctx.Step(`JwtMutatorHeader: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithHeaderMutator(_, header, headerValue string) error {
	s.manifestTemplate["header"] = header
	s.manifestTemplate["headerValue"] = headerValue
	return nil
}

func (s *istioJwtManifestScenario) shouldReturnResponseWithHeader(path, header, headerValue string) error {
	bodyContent := fmt.Sprintf(`"%s": "%s"`, header, headerValue)
	return s.callingTheEndpointWithValidTokenShouldResultInBodyContaining(path, "JWT", bodyContent)
}
