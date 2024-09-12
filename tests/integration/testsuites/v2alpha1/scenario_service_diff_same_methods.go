package v2alpha1

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"strings"
)

func initServiceDifferentSameMethods(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("service-diff-same-methods.yaml", "service-diff-same-methods")

	ctx.Step(`^ServiceDiffSvcSameMethods: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ServiceDiffSvcSameMethods: There is a workload and service for httpbin and helloworld$`, scenario.thereAreTwoServices)
	ctx.Step(`^ServiceDiffSvcSameMethods: There is an endpoint secured with JWT on path "([^"]*)" for httpbin service with methods '(\[.*\])'$`, scenario.thereIsAJwtSecuredPathWithMethods)
	ctx.Step(`^ServiceDiffSvcSameMethods: There is an endpoint secured with JWT on path "([^"]*)" for helloworld service with methods '(\[.*\])'$`, scenario.thereIsAJwtSecuredPathWithMethods)
	ctx.Step(`^ServiceDiffSvcSameMethods: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^ServiceDiffSvcSameMethods: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^ServiceDiffSvcSameMethods: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^ServiceDiffSvcSameMethods: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereAreTwoServices() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-app.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	_, err = s.resourceManager.CreateResources(s.k8sClient, resources...)
	return err
}

func (s *scenario) thereIsAJwtSecuredPathWithMethods(path string, methods string) {
	pathName := strings.TrimPrefix(path, "/")
	s.ManifestTemplate[fmt.Sprintf("%s%s", pathName, "Methods")] = methods
	s.ManifestTemplate[fmt.Sprintf("%sJwtSecuredPath", pathName)] = path
}
