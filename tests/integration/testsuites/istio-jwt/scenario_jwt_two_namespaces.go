package istiojwt

import (
	"github.com/cucumber/godog"
)

func initJwtTwoNamespaces(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("istio-jwt-two-namespaces.yaml", "istio-jwt-two-namespaces")

	ctx.Step(`^TwoNamespaces: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^TwoNamespaces: There are two namespaces with workload`, scenario.thereAreTwoNamespaces)
	ctx.Step(`^TwoNamespaces: There is an endpoint secured with JWT on path "([^"]*)" in APIRule Namespace$`, scenario.thereIsAnJwtSecuredPath)
	ctx.Step(`^TwoNamespaces: There is an endpoint secured with JWT on path "([^"]*)" in different namespace$`, scenario.thereIsAnJwtSecuredPathInDifferentNamespace)
	ctx.Step(`^TwoNamespaces: The APIRule is applied$`, scenario.theAPIRuleIsApplied)
	ctx.Step(`^TwoNamespaces: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^TwoNamespaces: Calling the "([^"]*)" endpoint without token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^TwoNamespaces: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *scenario) thereIsAnJwtSecuredPathInDifferentNamespace(path string) {
	s.ManifestTemplate["otherNamespacePath"] = path
}
