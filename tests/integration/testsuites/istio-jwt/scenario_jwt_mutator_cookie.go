package istiojwt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

func initMutatorCookie(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-mutator-cookie.yaml", "istio-jwt-mutator-cookie")

	ctx.Step(`^JwtMutatorCookie: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtMutatorCookie: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtMutatorCookie: Calling the "([^"]*)" endpoint should return response with cookie "([^"]*)" with value "([^"]*)"$`, scenario.shouldReturnResponseWithCookie)
	ctx.Step(`^JwtMutatorCookie: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) shouldReturnResponseWithCookie(path, cookie, cookieValue string) error {
	asserter := &helpers.BodyHasCookieValuePredicate{Expected: [][2]string{{cookie, cookieValue}}}
	tokenFrom := tokenFrom{
		From:     testcontext.AuthorizationHeaderName,
		Prefix:   testcontext.AuthorizationHeaderPrefix,
		AsHeader: true,
	}
	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(path, "/")), http.MethodGet, "JWT", asserter, nil, &tokenFrom)
}
