package ory

import (
	"github.com/cucumber/godog"
)

func initExposeMethodsOnPathsOAuth2Handler(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario("multiple-paths-and-methods-oauth2-handler.yaml", "multiple-paths-and-methods-oauth2-handler")

	ctx.Step(`^ExposeMethodsOnPathsOAuth2Handler: There is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^ExposeMethodsOnPathsOAuth2Handler: The APIRule is applied$`, scenario.theManifestIsApplied)
	ctx.Step(`^ExposeMethodsOnPathsOAuth2Handler: Calling the "([^"]*)" endpoint with "([^"]*)" method with a valid "([^"]*)" token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithMethodWithValidTokenShouldResultInStatusBetween)
	ctx.Step(`^ExposeMethodsOnPathsOAuth2Handler: Teardown httpbin service$`, scenario.teardownHttpbinService)
}
