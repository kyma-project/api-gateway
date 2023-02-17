package api_gateway

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"strings"
)

func initDiffServiceSameMethods(ctx *godog.ScenarioContext) {
	s, err := CreateScenarioWithRawAPIResource("istio-jwt-diff-svc-same-methods.yaml", "istio-diff-service-same-methods")
	if err != nil {
		t.Fatalf("could not initialize scenario err=%s", err)
	}

	scenario := istioJwtManifestScenario{s}

	ctx.Step(`DiffSvcSameMethods: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`DiffSvcSameMethods: There is a workload and service for httpbin and helloworld$`, scenario.thereAreTwoServices)
	ctx.Step(`DiffSvcSameMethods: There is an endpoint secured with JWT on path "([^"]*)" for httpbin service with methods '(\[.*\])'$`, scenario.thereIsAJwtSecuredPathWithMethods)
	ctx.Step(`DiffSvcSameMethods: There is an endpoint secured with JWT on path "([^"]*)" for helloworld service with methods '(\[.*\])'$`, scenario.thereIsAJwtSecuredPathWithMethods)
	ctx.Step(`DiffSvcSameMethods: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`DiffSvcSameMethods: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`DiffSvcSameMethods: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
}

func (s *istioJwtManifestScenario) thereAreTwoServices() error {
	resources, err := manifestprocessor.ParseFromFileWithTemplate("testing-helloworld-app.yaml", s.apiResourceDirectory, resourceSeparator, s.manifestTemplate)
	if err != nil {
		return err
	}
	_, err = batch.CreateResources(k8sClient, resources...)
	return err
}

func (s *istioJwtManifestScenario) thereIsAJwtSecuredPathWithMethods(path string, methods string) {
	pathName := strings.TrimPrefix(path, "/")
	s.manifestTemplate[fmt.Sprintf("%s%s", pathName, "Methods")] = methods
	s.manifestTemplate[fmt.Sprintf("%sJwtSecuredPath", pathName)] = path
}
