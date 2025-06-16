package ory

import (
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initCustomCors(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("custom-cors.yaml", "ory-custom-cors")

	ctx.Step(`^CustomCORS: There is an httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^CustomCORS: The APIRule with custom CORS setup is applied$`, scenario.applyApiRuleWithCustomCORS)
	ctx.Step(`^CustomCORS: Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and response header "([^"]*)" with value "([^"]*)"$`, scenario.preflightEndpointCallResponseHeaders)
	ctx.Step(`^CustomCORS: Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and no response header "([^"]*)"$`, scenario.preflightEndpointCallNoResponseHeader)
	ctx.Step(`^CustomCORS: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) applyApiRuleWithCustomCORS() error {
	r, err := manifestprocessor.ParseFromFileWithTemplate(s.ApiResourceManifestPath, s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}

	_, err = s.resourceManager.CreateOrUpdateResources(s.k8sClient, r...)
	if err != nil {
		return err
	}

	return nil
}
