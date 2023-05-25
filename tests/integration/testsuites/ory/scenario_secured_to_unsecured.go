package ory

import (
	_ "embed"
	"github.com/cucumber/godog"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"
)

type secureToUnsecureScenario struct {
	*scenario
}

func initSecuredToUnsecuredEndpoint(ctx *godog.ScenarioContext, ts *testsuite) {
	s := ts.createScenario("secured-to-unsecured.yaml", "secured-to-unsecured")

	scenario := secureToUnsecureScenario{s}

	ctx.Step(`^SecureToUnsecure: There is an httpbin application secured with OAuth2$`, scenario.thereIsAHttpbinServiceAndApiRuleIsApplied)
	ctx.Step(`^SecureToUnsecure: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^SecureToUnsecure: Update APIRule to expose the endpoint with noop strategy$`, scenario.updateApiRuleToMakeEndpointUnsecured)
	ctx.Step(`^SecureToUnsecure: Calling the "([^"]*)" endpoint with any token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^SecureToUnsecure: Calling the "([^"]*)" endpoint without a token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^SecureToUnsecure: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (s *secureToUnsecureScenario) updateApiRuleToMakeEndpointUnsecured() error {
	r, err := manifestprocessor.ParseFromFileWithTemplate("secured-to-unsecured-2.yaml", s.ApiResourceDirectory, s.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.ApplyApiRule(s.resourceManager.UpdateResources, s.resourceManager.UpdateResources, s.k8sClient, testcontext.GetRetryOpts(s.config), r)
}
