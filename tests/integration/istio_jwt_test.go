package api_gateway

import (
	_ "embed"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/jwt"
)

const (
	istioJwtApiRuleFile     string = "istio-jwt-strategy.yaml"
	happyPathManifestFile   string = "istio-jwt-scopes-happy.yaml"
	unhappyPathManifestFile string = "istio-jwt-scopes-unhappy.yaml"
	audiencesManifestFile   string = "istio-jwt-audiences.yaml"
)

type istioJwtUnstructuredScenario struct {
	*UnstructuredScenario
}

type istioJwtManifestScenario struct {
	*ScenarioWithRawAPIResource
}

func InitScenarioIstioJWT(ctx *godog.ScenarioContext) {
	initCommon(ctx)
	initScopesHappyPath(ctx)
	initScopesUnhappyPath(ctx)
	initAudience(ctx)
}

func initCommon(ctx *godog.ScenarioContext) {
	s, err := CreateScenario(istioJwtApiRuleFile, "istio-jwt")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtUnstructuredScenario{s}

	ctx.Step(`^Common: There is a deployment secured with JWT on path "([^"]*)"$`, scenario.thereIsAnEndpoint)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^Common: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
}

func initScopesHappyPath(ctx *godog.ScenarioContext) {
	s, err := CreateScenario(happyPathManifestFile, "istio-jwt-scopes-happy")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtUnstructuredScenario{s}

	ctx.Step(`^ScopesHappy: There is a deployment secured with JWT on path "([^"]*)"$`, scenario.thereIsAnEndpoint)
	ctx.Step(`^ScopesHappy: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with scopes read and write should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
}

func initScopesUnhappyPath(ctx *godog.ScenarioContext) {
	s, err := CreateScenario(unhappyPathManifestFile, "istio-jwt-scopes-unhappy")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtUnstructuredScenario{s}

	ctx.Step(`^ScopesUnhappy: There is a deployment secured with JWT on path "([^"]*)"$`, scenario.thereIsAnEndpoint)
	ctx.Step(`^ScopesUnhappy: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with scopes read and write should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
}

func (o *istioJwtUnstructuredScenario) thereIsAnEndpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, o.apiResource)
}

func (o *istioJwtUnstructuredScenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(path, tokenType string, lower, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.GetAccessToken(oauth2Cfg, jwtConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return godog.ErrUndefined
}

func (o *istioJwtUnstructuredScenario) callingTheEndpointWithInvalidTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}

func (o *istioJwtUnstructuredScenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", o.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
