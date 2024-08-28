package istiojwt

import (
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
)

func initCustomCors(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-custom-cors.yaml", "istio-custom-cors")

	ctx.Step(`^CustomCORS: There is an httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^CustomCORS: The APIRule with following CORS setup is applied AllowOrigins:'(\[.*\])', AllowMethods:'(\[.*\])', AllowHeaders:'(\[.*\])', AllowCredentials:"([^"]*)", ExposeHeaders:'(\[.*\])', MaxAge:"([^"]*)"$`, scenario.applyApiRuleWithCustomCORS)
	ctx.Step(`^CustomCORS: Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and response header "([^"]*)" with value "([^"]*)"$`, scenario.preflightEndpointCallResponseHeaders)
	ctx.Step(`^CustomCORS: Preflight calling the "([^"]*)" endpoint with header Origin:"([^"]*)" should result in status code (\d+) and no response header "([^"]*)"$`, scenario.preflightEndpointCallNoResponseHeader)
	ctx.Step(`^CustomCORS: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) applyApiRuleWithCustomCORS(allowOrigins, allowMethods, allowHeaders, allowCredentials, exposeHeaders, maxAge string) error {
	s.ManifestTemplate["AllowOrigins"] = allowOrigins
	s.ManifestTemplate["AllowMethods"] = allowMethods
	s.ManifestTemplate["AllowHeaders"] = allowHeaders
	s.ManifestTemplate["AllowCredentials"] = allowCredentials
	s.ManifestTemplate["ExposeHeaders"] = exposeHeaders
	s.ManifestTemplate["MaxAge"] = maxAge
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
