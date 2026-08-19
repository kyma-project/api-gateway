package istiojwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

func initMutatorHeader(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-mutator-header.yaml", "istio-jwt-mutator-header")

	ctx.Step(`^JwtMutatorHeader: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtMutatorHeader: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtMutatorHeader: Calling the "([^"]*)" endpoint should return response with header "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithHeader)
	ctx.Step(`^JwtMutatorHeader: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) shouldReturnResponseWithHeader(path, header, headerValue string) error {
	asserter := &helpers.BodyHasHeaderValuePredicate{Expected: [][2]string{{header, headerValue}}}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(path, "/")), http.MethodGet, "JWT", asserter, nil, &tokenFrom)
}
