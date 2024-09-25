package v2alpha1

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"net/http"
	"strings"
)

func initJwtFromParam(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("jwt-token-from-param.yaml", "jwt-token-from-param")
	scenario.ManifestTemplate["FromParamName"] = "jwt_token"

	ctx.Step(`^JwtTokenFromParam: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^JwtTokenFromParam: The APIRule is applied$`, scenario.theAPIRuleV2Alpha1IsApplied)
	ctx.Step(`^JwtTokenFromParam: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromParam: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from default header should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromParam: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token from parameter "([^"]*)" should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween)
	ctx.Step(`^JwtTokenFromParam: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) callingTheEndpointWithValidTokenFromParameterShouldResultInStatusBetween(endpoint, tokenType string, fromParameter string, lower, higher int) error {
	asserter := &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher}
	tokenFrom := tokenFrom{
		From:     fromParameter,
		AsHeader: false,
	}
	return s.callingEndpointWithMethodAndHeaders(fmt.Sprintf("%s/%s", s.Url, strings.TrimLeft(endpoint, "/")), http.MethodGet, tokenType, asserter, nil, &tokenFrom)
}
