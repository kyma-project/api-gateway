package api_gateway

import (
	"fmt"
	"github.com/cucumber/godog"
)

func initMutatorCookie(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-mutator-cookie.yaml", "istio-jwt-mutator-cookie")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`JwtMutatorCookie: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`JwtMutatorCookie: There is an endpoint on path "([^"]*)" with a cookie mutator setting "([^"]*)" cookie to "([^"]*)"$`, scenario.thereIsAnEndpointWithCookieMutator)
	ctx.Step(`JwtMutatorCookie: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`JwtMutatorCookie: Calling the "([^"]*)" endpoint should return response with cookie "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithCookie)
	ctx.Step(`JwtMutatorCookie: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithCookieMutator(_, header, headerValue string) error {
	s.manifestTemplate["cookie"] = header
	s.manifestTemplate["cookieValue"] = headerValue
	return nil
}

func (s *istioJwtManifestScenario) shouldReturnResponseWithCookie(path, cookie, cookieValue string) error {
	bodyContent := fmt.Sprintf(`"%s": "%s"`, cookie, cookieValue)
	return s.callingTheEndpointWithValidTokenShouldResultInBodyContaining(path, "JWT", bodyContent)
}
