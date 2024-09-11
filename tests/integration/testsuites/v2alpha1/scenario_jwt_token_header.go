package v2alpha1

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"net/http"
	"strings"
)

func initJwtFromHeader(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-token-from-header.yaml", "jwt-token-from-header")
	scenario.ManifestTemplate["JWTHeaderName"] = "x-jwt-token"

	ctx.Step(`^JwtTokenFromHeader: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtTokenFromHeader: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^JwtTokenFromHeader: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromHeader: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from default header should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromHeader: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from header "([^"]*)" and prefix "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromHeader: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) callingTheEndpointWithValidTokenFromHeaderShouldResultInStatusBetween(endpoint, tokenType string, fromHeader string, prefix string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromHeader,
		Prefix:   prefix,
		AsHeader: true,
	}
	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, tokenType, asserter, nil, &tokenFrom)
}
