package istiojwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

func initMultipleMutators(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-multiple-mutators.yaml", "istio-jwt-multiple-mutators")

	ctx.Step(`^JwtMultipleMutators: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(
		`^JwtMultipleMutators: There is an endpoint on path "([^"]*)" with a header mutator setting "([^"]*)" header to "([^"]*)" and "([^"]*)" header to "([^"]*)"$`,
		scenario.thereIsAnEndpointWithHeaderMutatorWithTwoHeaders,
	)
	ctx.Step(
		`^JwtMultipleMutators: There is an endpoint on path "([^"]*)" with a cookie mutator setting "([^"]*)" cookie to "([^"]*)" and "([^"]*)" cookie to "([^"]*)"$`,
		scenario.thereIsAnEndpointWithCookieMutatorWithTwoCookies,
	)
	ctx.Step(`^JwtMultipleMutators: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^JwtMultipleMutators: Calling the "([^"]*)" endpoint should return response with cookies "([^"]*)" with value "([^"]*)" and "([^"]*)" with value "([^"]*)"$`,
		scenario.shouldReturnResponseWithKeyValuePairs,
	)
	ctx.Step(
		`^JwtMultipleMutators: Calling the "([^"]*)" endpoint should return response with headers "([^"]*)" with value "([^"]*)" and "([^"]*)" with value "([^"]*)"$`,
		scenario.shouldReturnResponseWithKeyValuePairs,
	)
	ctx.Step(`^JwtMultipleMutators: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnEndpointWithHeaderMutatorWithTwoHeaders(_, header1, header1Value, header2, header2Value string) error {
	s.ManifestTemplate["header1"] = header1
	s.ManifestTemplate["header1Value"] = header1Value
	s.ManifestTemplate["header2"] = header2
	s.ManifestTemplate["header2Value"] = header2Value
	return nil
}

func (s *scenario) thereIsAnEndpointWithCookieMutatorWithTwoCookies(_, cookie1, cookie1Value, cookie2, cookie2Value string) error {
	s.ManifestTemplate["cookie1"] = cookie1
	s.ManifestTemplate["cookie1Value"] = cookie1Value
	s.ManifestTemplate["cookie2"] = cookie2
	s.ManifestTemplate["cookie2Value"] = cookie2Value
	return nil
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
