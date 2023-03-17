package api_gateway

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
)

func initMutatorHeader(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-mutator-header.yaml", "istio-mutator-header")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`Mutator-Header: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`Mutator-Header: There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)"$`, scenario.thereIsAnEndpointWithHeaderMutator)
	ctx.Step(`Mutator-Header: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Mutator-Header: Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithHeader)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithHeaderMutator(_, header, headerValue string) error {
	s.manifestTemplate["header"] = header
	s.manifestTemplate["headerValue"] = headerValue
	return nil
}

func (s *istioJwtManifestScenario) shouldReturnResponseWithHeader(path, header, headerValue string) error {
	bodyContent := fmt.Sprintf(`"%s": "%s"`, header, headerValue)
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", s.url, path), &helpers.BodyContainsPredicate{Expected: bodyContent})
}
