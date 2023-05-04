package api_gateway

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initCustomLabelSelector(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-custom-label-selector.yaml", "istio-jwt-custom-label-selector")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`CustomLabelSelector: There is a helloworld service with custom label selector name "([^"]*)"$`, scenario.thereIsHelloworldCustomLabelService)
	ctx.Step(`CustomLabelSelector: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`CustomLabelSelector: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`CustomLabelSelector: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`CustomLabelSelector: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`CustomLabelSelector: Teardown helloworld service$`, scenario.teardownHelloworldCustomLabelService)
}

func (s *istioJwtManifestScenario) thereIsHelloworldCustomLabelService(labelName string) error {
	s.manifestTemplate["CustomLabelName"] = labelName
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-custom-label-app.yaml", s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	_, err = batch.CreateResources(k8sClient, resources...)

	s.url = fmt.Sprintf("https://helloworld-%s.%s", s.testID, s.domain)

	return err
}

func (s *istioJwtManifestScenario) teardownHelloworldCustomLabelService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-custom-label-app.yaml", s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	err = batch.DeleteResources(k8sClient, resources...)
	if err != nil {
		return err
	}

	s.url = ""

	return nil
}
