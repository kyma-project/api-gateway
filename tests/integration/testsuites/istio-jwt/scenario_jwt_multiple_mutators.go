package istiojwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
)

func initMultipleMutators(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-multiple-mutators.yaml", "istio-jwt-multiple-mutators")

	ctx.Step(`^JwtMultipleMutators: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtMultipleMutators: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtMultipleMutators: Calling the "([^"]*)" endpoint should return response with cookies "([^"]*)" with value "([^"]*)" and "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithKeyValuePairs)
	ctx.Step(`^JwtMultipleMutators: Calling the "([^"]*)" endpoint should return response with headers "([^"]*)" with value "([^"]*)" and "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithKeyValuePairs)
	ctx.Step(`^JwtMultipleMutators: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) shouldReturnResponseWithKeyValuePairs(endpoint, k1, v1, k2, v2 string) error {
	expectedInBody := []string{
		fmt.Sprintf(`"%s": "%s"`, k1, v1),
		fmt.Sprintf(`"%s": "%s"`, k2, v2),
	}

	asserter := &helpers.BodyContainsPredicate{Expected: expectedInBody}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}

	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, "JWT", asserter, nil, &tokenFrom)
}
