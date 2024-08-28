package v2alpha1

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initServiceCustomLabelSelector(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("service-custom-label-selector.yaml", "service-custom-label-selector")

	ctx.Step(`^ServiceCustomLabelSelector: There is a helloworld service with custom label selector name "([^"]*)"$`, scenario.thereIsHelloworldCustomLabelService)
	ctx.Step(`^ServiceCustomLabelSelector: There is an endpoint secured with JWT on path "([^"]*)"`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^ServiceCustomLabelSelector: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^ServiceCustomLabelSelector: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^ServiceCustomLabelSelector: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^ServiceCustomLabelSelector: Teardown helloworld service$`, scenario.teardownHelloworldCustomLabelService)
}

func (s *scenario) thereIsHelloworldCustomLabelService(labelName string) error {
	s.ManifestTemplate["CustomLabelName"] = labelName
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-custom-label-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)

	s.Url = fmt.Sprintf("https://helloworld-%s.%s", s.TestID, s.Domain)

	return err
}

func (s *scenario) teardownHelloworldCustomLabelService() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-custom-label-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	err = s.resourceManager.DeleteResources(s.k8sClient, resources...)
	if err != nil {
		return err
	}

	s.Url = ""

	return nil
}
