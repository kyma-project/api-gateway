package api_gateway

import (
	_ "embed"
	"fmt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"

	"github.com/cucumber/godog"
)

type istioJwtManifestScenario struct {
	*ScenarioWithRawAPIResource
}

func InitScenarioIstioJWT(ctx *godog.ScenarioContext) {
	initCommon(ctx)
	initRequiredScopes(ctx)
	initAudience(ctx)
	initJwtAndAllow(ctx)
}

func (s *istioJwtManifestScenario) theAPIRuleIsApplied() error {
	resource, err := manifestprocessor.ParseFromFileWithTemplate(s.apiResourceManifestPath, s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, resource)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithAValidJWTToken(path, tokenType, _, _ string, lower, higher int) error {
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, lower, higher)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithValidTokenShouldResultInStatusBetween(path, tokenType string, lower, higher int) error {
	return callingEndpointWithHeadersWithRetries(s.url, path, tokenType, lower, higher)
}

func callingEndpointWithHeadersWithRetries(url string, path string, tokenType string, lower int, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.GetAccessToken(oauth2Cfg, jwtConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return godog.ErrUndefined
}

func (s *istioJwtManifestScenario) callingTheEndpointWithoutTokenShouldResultInStatusBetween(path string, lower, higher int) error {
	return helper.CallEndpointWithRetries(fmt.Sprintf("%s%s", s.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
}
