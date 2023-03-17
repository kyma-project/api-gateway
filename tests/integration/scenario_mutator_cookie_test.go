package api_gateway

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
)

func initMutatorCookie(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-mutator-cookie.yaml", "istio-mutator-cookie")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Mutator-Cookie: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`Mutator-Cookie: There is an endpoint on path "([^"]*)" with a cookies mutator setting "([^"]*)" cookie to "([^"]*)"$`, scenario.thereIsAnEndpointWithCookieMutator)
	ctx.Step(`Mutator-Cookie: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Mutator-Cookie: Calling the "([^"]*)" endpoint should return response with cookie "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithCookie)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithCookieMutator(_, header, headerValue string) error {
	s.manifestTemplate["cookie"] = header
	s.manifestTemplate["cookieValue"] = headerValue
	return nil
}

func (s *istioJwtManifestScenario) shouldReturnResponseWithCookie(path, cookie, cookieValue string) error {
	bodyContent := fmt.Sprintf(`"%s": "%s"`, cookie, cookieValue)
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", s.url, path), &helpers.BodyContainsPredicate{Expected: bodyContent})
}
