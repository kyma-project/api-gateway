package istiojwt

import (
	"fmt"

	"github.com/cucumber/godog"
)

func initMutatorHeader(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-mutator-header.yaml", "istio-jwt-mutator-header")

	ctx.Step(`^JwtMutatorHeader: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtMutatorHeader: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtMutatorHeader: Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithHeader)
	ctx.Step(`^JwtMutatorHeader: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithHeaderMutator(_, header, headerValue string) error {
	s.ManifestTemplate["header"] = header
	s.ManifestTemplate["headerValue"] = headerValue
	return nil
}

func (s *scenario) shouldReturnResponseWithHeader(path, header, headerValue string) error {
	bodyContent := fmt.Sprintf(`"%s": "%s"`, header, headerValue)
	return s.callingTheEndpointWithValidTokenShouldResultInBodyContaining(path, "JWT", bodyContent)
}
