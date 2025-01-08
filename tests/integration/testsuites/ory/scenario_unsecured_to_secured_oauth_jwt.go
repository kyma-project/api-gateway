package ory

import (
	_ "embed"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/manifestprocessor"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/testcontext"

	"github.com/cucumber/godog"
)

type unsecuredToSecured struct {
	*scenario
}

func initUnsecuredToSecured(ctx *godog.ScenarioContext, ts *testsuite) {
	s := ts.createScenario("unsecured-to-secured.yaml", "unsecured-to-secured")
	scenario := unsecuredToSecured{s}

	ctx.Step(`^UnsecureToSecure: There is an unsecured API with all paths available without authorization$`, scenario.thereIsAHttpbinServiceAndApiRuleIsApplied)
	ctx.Step(`^UnsecureToSecure: Calling the "([^"]*)" endpoint without a token should result in status beetween (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^UnsecureToSecure: API is secured with OAuth2 on path \/headers and JWT on path \/image$`, scenario.secureWithOAuth2JWT)
	ctx.Step(`^UnsecureToSecure: Calling the "([^"]*)" endpoint with an invalid token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithInvalidTokenShouldResultInStatusBetween)
	ctx.Step(`^UnsecureToSecure: Calling the "([^"]*)" endpoint with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^UnsecureToSecure: Calling the "([^"]*)" endpoint without a token should result in status between (\d+) and (\d+)$$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusBetween)
	ctx.Step(`^UnsecureToSecure: Teardown httpbin service$`, scenario.teardownHttpbinService)
}

func (u *unsecuredToSecured) secureWithOAuth2JWT() error {
	r, err := manifestprocessor.ParseSingleEntryFromFileWithTemplate("unsecured-to-secured-2.yaml", u.ApiResourceDirectory, u.ManifestTemplate)
	if err != nil {
		return err
	}
	return helpers.UpdateApiRule(u.resourceManager, u.k8sClient, testcontext.GetRetryOpts(), r)
}
