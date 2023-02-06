package api_gateway

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/jwt"
	"github.com/kyma-incubator/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initAudience(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource(audiencesManifestFile, "istio-jwt-audiences")
	if err != nil {
		t.Fatalf("could not initialize unsecure endpoint scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`^Audiences: There is a endpoint secured with JWT on path "([^"]*)" requiring audiences '(\[.*\])'$`, scenario.thereIsAnEndpointWithAudiences)
	ctx.Step(`^Audiences: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`Audiences: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token with audiences "([^"]*)" and "([^"]*)" should result in status between (\d+) and (\d+)`, scenario.callingTheEndpointWithAValidJWTToken)
}

func (s *istioJwtManifestScenario) thereIsAnEndpointWithAudiences(path string, audiences string) error {
	s.manifestTemplate[fmt.Sprintf("%s%s", strings.TrimPrefix(path, "/"), "Audiences")] = audiences
	return nil
}

func (s *istioJwtManifestScenario) theAPIRuleIsApplied() error {
	resource, err := manifestprocessor.ParseFromFileWithTemplate(s.apiResourceManifestPath, s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, resource)
}

func (s *istioJwtManifestScenario) callingTheEndpointWithAValidJWTToken(path, tokenType, _, _ string, lower, higher int) error {
	switch tokenType {
	case "JWT":
		tokenJWT, err := jwt.GetAccessToken(oauth2Cfg, jwtConfig)
		if err != nil {
			return fmt.Errorf("failed to fetch an id_token: %s", err.Error())
		}
		headerVal := fmt.Sprintf("Bearer %s", tokenJWT)

		return helper.CallEndpointWithHeadersWithRetries(headerVal, authorizationHeaderName, fmt.Sprintf("%s%s", s.url, path), &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
	}
	return godog.ErrUndefined
}
