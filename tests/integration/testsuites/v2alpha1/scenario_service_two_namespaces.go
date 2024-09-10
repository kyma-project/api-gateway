package v2alpha1

import (
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initServiceTwoNamespaces(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("service-two-namespaces.yaml", "service-two-namespaces")

	ctx.Step(`^ServiceTwoNamespaces: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ServiceTwoNamespaces: There is a service with workload in a second namespace`, scenario.thereIsServiceInSecondNamespace)
	ctx.Step(`^ServiceTwoNamespaces: There is an endpoint secured with JWT on path "([^"]*)" in APIRule Namespace$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^ServiceTwoNamespaces: There is an endpoint secured with JWT on path "([^"]*)" in different namespace$`, scenario.thereIsAnJwtSecuredPathInDifferentNamespace)
	ctx.Step(`^ServiceTwoNamespaces: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^ServiceTwoNamespaces: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^ServiceTwoNamespaces: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^ServiceTwoNamespaces: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnJwtSecuredPathInDifferentNamespace(path string) {
	s.ManifestTemplate["otherNamespacePath"] = path
}

func (s *scenario) thereIsServiceInSecondNamespace() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("second-service.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	return err
}
