package ratelimit

import (
	"github.com/cucumber/godog"
)

func initScenario(ctx *godog.ScenarioContext, ts *testsuite) {
	scenario := ts.createScenario()

	ctx.Step(`^calling the "([^"]*)" endpoint should result in status code (\d+) for requests$`, scenario.callingEndpointNTimesShouldResultWithStatusCode)
	ctx.Step(`^calling the "([^"]*)" endpoint with header should result in status code (\d+) for requests$`, scenario.callingEndpointWithHeadersNTimesShouldResultWithStatusCode)
	ctx.Step(`^RateLimit path-based configuration is applied$`, scenario.rateLimitWithPathBaseConfigurationApplied)
	ctx.Step(`^RateLimit header-based configuration is applied$`, scenario.rateLimitWithHeaderBaseConfigurationApplied)
	ctx.Step(`^RateLimit path and header based configuration is applied$`, scenario.rateLimitWithPathAndHeaderBaseConfigurationApplied)
	ctx.Step(`^there is a httpbin service$`, scenario.thereIsAHttpbinService)
	ctx.Step(`^Teardown httpbin service$`, scenario.teardownHttpbinService)
	ctx.Step(`^RateLimit targeting Istio ingress gateway with path-based configuration is applied$`, scenario.rateLimitTargetingIngressWithPathBaseConfigurationApplied)
	ctx.Step(`^RateLimit targeting Istio ingress gateway with header-based configuration is applied$`, scenario.rateLimitTargetingIngressWithHeaderBaseConfigurationApplied)
	ctx.Step(`^RateLimit targeting Istio ingress gateway with path and header based configuration is applied$`, scenario.rateLimitTargetingIngressWithPathAndHeaderBaseConfigurationApplied)
	ctx.Step(`^Teardown RateLimit`, scenario.tearDownRateLimit)
}
