package istiojwt

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initDiffServiceSameMethods(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-diff-svc-same-methods.yaml", "istio-diff-service-same-methods")

	ctx.Step(`^DiffSvcSameMethods: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^DiffSvcSameMethods: There is a workload and service for httpbin and helloworld$`, scenario.thereAreTwoServices)
	ctx.Step(
		`^DiffSvcSameMethods: There is an endpoint secured with JWT on path "([^"]*)" for httpbin service with methods '(\[.*\])'$`,
		scenario.thereIsAJwtSecuredPathWithMethods,
	)
	ctx.Step(
		`^DiffSvcSameMethods: There is an endpoint secured with JWT on path "([^"]*)" for helloworld service with methods '(\[.*\])'$`,
		scenario.thereIsAJwtSecuredPathWithMethods,
	)
	ctx.Step(`^DiffSvcSameMethods: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(
		`^DiffSvcSameMethods: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween,
	)
	ctx.Step(
		`^DiffSvcSameMethods: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`,
		scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween,
	)
	ctx.Step(`^DiffSvcSameMethods: Teardown httpbin service$`, scenario.teardownHttpbinService)
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
