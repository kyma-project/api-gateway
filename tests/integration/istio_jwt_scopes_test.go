package api_gateway

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/jwt"
)

type istioJwtScopesScenario struct {
	*Scenario
}

func InitializeScenarioIstioJWTScopes(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateScenario(istioJwtApiRuleScopesFile, "istio-jwt-scopes")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtScopesScenario{mainScenario}

	ctx.Step(`^IstioJWTScopes: There is a deployment secured with JWT on path "([^"]*)"$`, scenario.thereIsAnEndpoint)
	ctx.Step(`^IstioJWTScopes: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^IstioJWTScopes: Calling the "([^"]*)" endpoint with a invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^IstioJWTScopes: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token and valid scopes should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidScopesShouldResultInStatusBetween)
	ctx.Step(`^IstioJWTScopes: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token and invalid scopes should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenWithoutScopesShouldResultInStatusBetween)
}

func (o *istioJwtScopesScenario) thereIsAnEndpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, o.apiResource)
}

func (o *istioJwtScopesScenario) callingTheEndpointWithValidScopesShouldResultInStatusBetween(path, tokenType string, lower, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.GetAccessToken(oauth2Cfg, jwtConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return errors.New("should not happen")
}

func (o *istioJwtScopesScenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *istioJwtScopesScenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *istioJwtScopesScenario) callingTheEndpointWithValidTokenWithoutScopesShouldResultInStatusBetween(path, tokenType string, lower, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.GetAccessToken(oauth2Cfg, jwtConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return errors.New("should not happen")
}
