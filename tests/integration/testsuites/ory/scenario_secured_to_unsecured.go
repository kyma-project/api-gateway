package ory

import (
	_ "embed"
	"github.com/cucumber/godog"
)

type secureToUnsecureScenario struct {
	*scenario
}

func initScenarioSecuredToUnsecuredEndpoint(ctx *godog.ScenarioContext, ts *testsuite) {
	s := ts.createScenario("secured-to-unsecured.yaml", "secured-to-unsecured")

	scenario := secureToUnsecureScenario{s}

	ctx.Step(`^SecureToUnsecure: There is an endpoint secured with OAuth2$`, scenario.thereIsAnOauth2Endpoint)
	ctx.Step(`^SecureToUnsecure: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	//ctx.Step(`^SecureToUnsecure: Endpoint is exposed with noop strategy$`, scenario.unsecureTheEndpoint)
	//ctx.Step(`^SecureToUnsecure: Calling the endpoint with any token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusBeetween)
	//ctx.Step(`^SecureToUnsecure: Calling the endpoint without a token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutATokenShouldResultInStatusBeetween)
	//ctx.Step(`^SecureToUnsecure: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

// TODO
//func (u *secureToUnsecureScenario) unsecureTheEndpoint() error {
//	return helper.APIRuleWithRetries(batch.UpdateResources, batch.UpdateResources, k8sClient, u.apiResourceTwo)
//}
//
//func (u *secureToUnsecureScenario) callingTheEndpointWithAnyTokenShouldResultInStatusBeetween(lower int, higher int) error {
//	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, u.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
//}
//
//func (u *secureToUnsecureScenario) callingTheEndpointWithoutATokenShouldResultInStatusBeetween(lower int, higher int) error {
//	return helper.CallEndpointWithRetries(u.url, &helpers.StatusPredicate{LowerStatusBound: lower, UpperStatusBound: higher})
//}
